package api

import (
	_ "embed"
	"errors"
	"html/template"
	"net/http"
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

// InfoHandler wires the generated OpenAPI handlers with auxiliary endpoints
// that expose build information, status checks, and a pre-built HTML UI.
type InfoHandler struct {
	*Responder
	baseURL         string
	infoProvider    InfoProvider
	swaggerProvider SwaggerProvider
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

// GetStatus returns a simple health payload that can be used for probes and
// readiness checks.
func (ih *InfoHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	ih.RespondWithJSON(w, r, http.StatusOK, map[string]string{"status": "HEALTHY"})
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
