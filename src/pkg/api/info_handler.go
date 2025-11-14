package api

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

//go:embed embedded/stoplight.html
var openapiHTMLStoplight []byte

// InfoProvider returns the payload that will be exposed by the version endpoint.
// The provider allows callers to inject their own source for build metadata or
// runtime diagnostics.
type InfoProvider func() any

// SwaggerProvider returns the raw OpenAPI document that should be rendered by
// the documentation endpoints. It is commonly backed by an embedded JSON file
// generated at build time.
type SwaggerProvider func() ([]byte, error)

// InfoOption follows the functional options pattern used by NewInfoHandler to
// configure optional collaborators such as the responder, base URL, and
// information providers.
type InfoOption func(*InfoHandler)

const defaultProbeTimeout = 2 * time.Second

// ProbeFunc is executed to determine the outcome of liveness or readiness
// probes. Returning a non-nil error marks the probe as failed.
type ProbeFunc func(ctx context.Context) error

type probePayload struct {
	Status  string   `json:"status"`
	Details []string `json:"details,omitempty"`
}

// InfoHandler wires the generated OpenAPI handlers with auxiliary endpoints
// that expose build information, status checks, and a pre-built HTML UI.
type InfoHandler struct {
	*Responder
	baseURL         string
	infoProvider    InfoProvider
	swaggerProvider SwaggerProvider
	probeTimeout    time.Duration
	livenessChecks  []ProbeFunc
	readinessChecks []ProbeFunc
}

// NewInfoHandler constructs an InfoHandler with sensible defaults. Callers can
// supply InfoOption values to plug in domain specific providers or override the
// base responder implementation.
func NewInfoHandler(opts ...InfoOption) *InfoHandler {
	ih := &InfoHandler{
		Responder: NewResponder(),
		infoProvider: func() any {
			return map[string]string{}
		},
		swaggerProvider: func() ([]byte, error) {
			return nil, errors.New("api swagger provider not configured")
		},
		probeTimeout: defaultProbeTimeout,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(ih)
		}
	}
	return ih
}

// WithInfoResponder replaces the responder used to craft JSON responses and
// handle error reporting.
func WithInfoResponder(responder *Responder) InfoOption {
	return func(ih *InfoHandler) {
		if responder != nil {
			ih.Responder = responder
		}
	}
}

// WithBaseURL sets the URL prefix that will be injected into the rendered
// documentation template.
func WithBaseURL(baseURL string) InfoOption {
	return func(ih *InfoHandler) {
		ih.baseURL = baseURL
	}
}

// WithInfoProvider swaps the default metadata provider with a user supplied
// implementation.
func WithInfoProvider(provider InfoProvider) InfoOption {
	return func(ih *InfoHandler) {
		if provider != nil {
			ih.infoProvider = provider
		}
	}
}

// WithSwaggerProvider sets the source of the OpenAPI JSON document that backs
// the documentation endpoints.
func WithSwaggerProvider(provider SwaggerProvider) InfoOption {
	return func(ih *InfoHandler) {
		if provider != nil {
			ih.swaggerProvider = provider
		}
	}
}

// WithProbeTimeout adjusts the maximum duration allowed for probe checks.
func WithProbeTimeout(timeout time.Duration) InfoOption {
	return func(ih *InfoHandler) {
		if timeout > 0 {
			ih.probeTimeout = timeout
		}
	}
}

// WithLivenessChecks replaces the default liveness checks with the supplied
// functions.
func WithLivenessChecks(checks ...ProbeFunc) InfoOption {
	return func(ih *InfoHandler) {
		ih.livenessChecks = filterProbes(checks)
	}
}

// WithReadinessChecks replaces the default readiness checks with the supplied
// functions.
func WithReadinessChecks(checks ...ProbeFunc) InfoOption {
	return func(ih *InfoHandler) {
		ih.readinessChecks = filterProbes(checks)
	}
}

// GetStatus returns a simple health payload that can be used for lightweight
// diagnostics.
func (ih *InfoHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ih.respondProbe(w, r, http.StatusOK, "HEALTHY")
}

// GetHealthz implements the liveness probe recommended for Kubernetes.
func (ih *InfoHandler) GetHealthz(w http.ResponseWriter, r *http.Request) {
	if err := ih.runChecks(r.Context(), ih.livenessChecks); err != nil {
		ih.HandleAPIError(w, r, http.StatusServiceUnavailable, err, "liveness probe failed")
		return
	}
	ih.respondProbe(w, r, http.StatusOK, "ok")
}

// GetReadyz implements the readiness probe recommended for Kubernetes.
func (ih *InfoHandler) GetReadyz(w http.ResponseWriter, r *http.Request) {
	if err := ih.runChecks(r.Context(), ih.readinessChecks); err != nil {
		ih.HandleAPIError(w, r, http.StatusServiceUnavailable, err, "readiness probe failed")
		return
	}
	ih.respondProbe(w, r, http.StatusOK, "ready")
}

// GetVersion returns the structure provided by the configured InfoProvider.
func (ih *InfoHandler) GetVersion(w http.ResponseWriter, r *http.Request) {
	payload := ih.infoProvider()
	if payload == nil {
		payload = map[string]string{}
	}
	ih.RespondWithJSON(w, r, http.StatusOK, payload)
}

// GetOpenAPIJSON streams the configured OpenAPI JSON document to the caller.
func (ih *InfoHandler) GetOpenAPIJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	bytes, err := ih.swaggerProvider()
	if err != nil {
		ih.HandleInternalServerError(w, r, err, "failed to load swagger spec")
		return
	}

	if _, err = w.Write(bytes); err != nil {
		ih.HandleInternalServerError(w, r, err, "failed to write swagger response")
		return
	}
}

// GetOpenAPIHTML renders an embedded Stoplight viewer that fetches the OpenAPI
// document from the GetOpenAPIJSON endpoint.
func (ih *InfoHandler) GetOpenAPIHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	tmpl, err := template.New("openapi").Parse(string(openapiHTMLStoplight))
	if err != nil {
		ih.HandleInternalServerError(w, r, err, "failed to parse openapi template")
		return
	}

	if err := tmpl.Execute(w, map[string]string{
		"BaseURL": ih.baseURL,
	}); err != nil {
		ih.HandleInternalServerError(w, r, err, "failed to render openapi template")
		return
	}
}

func (ih *InfoHandler) respondProbe(w http.ResponseWriter, r *http.Request, statusCode int, state string, details ...string) {
	payload := probePayload{Status: state}
	if len(details) > 0 {
		payload.Details = append(payload.Details, details...)
	}
	ih.RespondWithJSON(w, r, statusCode, payload)
}

func (ih *InfoHandler) runChecks(ctx context.Context, checks []ProbeFunc) error {
	if len(checks) == 0 {
		return nil
	}

	timeout := ih.probeTimeout
	if timeout <= 0 {
		timeout = defaultProbeTimeout
	}

	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for idx, check := range checks {
		if check == nil {
			continue
		}

		if err := check(probeCtx); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return fmt.Errorf("probe %d timed out after %s", idx+1, timeout)
			}
			if errors.Is(err, context.Canceled) {
				return fmt.Errorf("probe %d was cancelled", idx+1)
			}
			return fmt.Errorf("probe %d failed: %w", idx+1, err)
		}
	}

	return nil
}

func filterProbes(checks []ProbeFunc) []ProbeFunc {
	if len(checks) == 0 {
		return nil
	}

	filtered := make([]ProbeFunc, 0, len(checks))
	for _, check := range checks {
		if check != nil {
			filtered = append(filtered, check)
		}
	}

	if len(filtered) == 0 {
		return nil
	}

	return filtered
}
