package apihandler

import (
	_ "embed"
	"errors"
	"html/template"
	"net/http"
	"strings"

	"drblury/event-driven-service/internal/domain"
	generator "drblury/event-driven-service/internal/server/gen"
	"drblury/event-driven-service/internal/usecase"
	"log/slog"

	infohandler "github.com/drblury/apiweaver/info"
	"github.com/drblury/apiweaver/responder"
)

//go:embed asyncapi.json
var asyncAPISpec []byte

// Supported UI types for API documentation
const (
	UIStoplight = "stoplight"
	UIScalar    = "scalar"
	UISwaggerUI = "swagger"
	UIRedoc     = "redoc"
)

// Embedded HTML templates for different OpenAPI UI types
var (
	stoplightTemplate = template.Must(template.New("stoplight").Parse(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>API Documentation - Stoplight</title>
    <script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
    <link rel="stylesheet" href="https://unpkg.com/@stoplight/elements/styles.min.css">
  </head>
  <body>
    <elements-api
      apiDescriptionUrl="{{ .SpecURL }}"
      router="hash"
      layout="sidebar"
    />
  </body>
</html>`))

	scalarTemplate = template.Must(template.New("scalar").Parse(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>API Documentation - Scalar</title>
  </head>
  <body>
    <script id="api-reference" data-url="{{ .SpecURL }}"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@latest"></script>
  </body>
</html>`))

	swaggerUITemplate = template.Must(template.New("swaggerui").Parse(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>API Documentation - Swagger UI</title>
    <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@latest/swagger-ui.css">
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@latest/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@latest/swagger-ui-standalone-preset.js"></script>
    <script>
      window.onload = function() {
        window.ui = SwaggerUIBundle({
          url: "{{ .SpecURL }}",
          dom_id: '#swagger-ui',
          deepLinking: true,
          presets: [
            SwaggerUIBundle.presets.apis,
            SwaggerUIStandalonePreset
          ],
          plugins: [
            SwaggerUIBundle.plugins.DownloadUrl
          ],
          layout: "StandaloneLayout"
        });
      };
    </script>
  </body>
</html>`))

	redocTemplate = template.Must(template.New("redoc").Parse(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>API Documentation - Redoc</title>
  </head>
  <body>
    <redoc spec-url="{{ .SpecURL }}"></redoc>
    <script src="https://cdn.jsdelivr.net/npm/redoc@latest/bundles/redoc.standalone.js"></script>
  </body>
</html>`))
)

type APIHandler struct {
	*infohandler.InfoHandler
	AppLogic        *usecase.AppLogic
	log             *slog.Logger
	baseURL         string
	uiHandlers      map[string]*infohandler.InfoHandler
	asyncAPIHandler *infohandler.InfoHandler
}

func NewAPIHandler(
	appLogic *usecase.AppLogic,
	info *domain.Info,
	logger *slog.Logger,
	baseURL string,
	docsTemplatePath string,
) *APIHandler {
	resp := responder.NewResponder(
		responder.WithLogger(logger),
		responder.WithErrorClassifier(func(err error) (int, bool) {
			switch {
			case errors.Is(err, domain.ErrorUpstreamService):
				return http.StatusInternalServerError, true
			case errors.Is(err, domain.ErrorNotFound), errors.Is(err, domain.ErrorBadRequest):
				return http.StatusBadRequest, true
			default:
				var validationErr domain.ErrValidations
				if errors.As(err, &validationErr) {
					return http.StatusBadRequest, true
				}
			}
			return 0, false
		}),
	)

	swaggerProvider := func() ([]byte, error) {
		swagger, err := generator.GetSwagger()
		if err != nil {
			return nil, err
		}
		return swagger.MarshalJSON()
	}

	infoProvider := func() any {
		return info
	}

	// Map of UI type to template
	uiTemplates := map[string]*template.Template{
		UIStoplight: stoplightTemplate,
		UIScalar:    scalarTemplate,
		UISwaggerUI: swaggerUITemplate,
		UIRedoc:     redocTemplate,
	}

	// Check for custom template override
	customTmpl := parseDocsTemplate(docsTemplatePath, logger)

	// Create handlers for each UI type
	uiHandlers := make(map[string]*infohandler.InfoHandler)
	for key, tmpl := range uiTemplates {
		templateToUse := tmpl
		if customTmpl != nil {
			templateToUse = customTmpl // Custom template overrides all UI types
		}

		options := []infohandler.InfoOption{
			infohandler.WithInfoResponder(resp),
			infohandler.WithBaseURL(baseURL),
			infohandler.WithInfoProvider(infoProvider),
			infohandler.WithSwaggerProvider(swaggerProvider),
			infohandler.WithOpenAPITemplate(templateToUse),
			infohandler.WithOpenAPITemplateData(func(_ *http.Request, base string) any {
				return map[string]any{
					"BaseURL": base,
					"SpecURL": openAPISpecURL(base),
				}
			}),
		}

		uiHandlers[key] = infohandler.NewInfoHandler(options...)
	}

	// Default handler (Stoplight)
	defaultHandler := uiHandlers[UIStoplight]

	// Create AsyncAPI handler using apiweaver's built-in support
	asyncAPIHandler := infohandler.NewInfoHandler(
		infohandler.WithInfoResponder(resp),
		infohandler.WithBaseURL(baseURL),
		infohandler.WithAsyncAPIProvider(func() ([]byte, error) {
			return asyncAPISpec, nil
		}),
	)

	return &APIHandler{
		InfoHandler:     defaultHandler,
		AppLogic:        appLogic,
		log:             logger,
		baseURL:         baseURL,
		uiHandlers:      uiHandlers,
		asyncAPIHandler: asyncAPIHandler,
	}
}

// GetOpenAPIHTML serves the OpenAPI documentation with UI selection via query parameter.
// Use ?ui=stoplight, ?ui=scalar, ?ui=swagger, or ?ui=redoc to select the UI.
// Defaults to Stoplight if no query parameter is provided.
func (h *APIHandler) GetOpenAPIHTML(w http.ResponseWriter, r *http.Request) {
	uiType := r.URL.Query().Get("ui")
	if uiType == "" {
		uiType = UIStoplight // default
	}

	handler, ok := h.uiHandlers[strings.ToLower(uiType)]
	if !ok {
		// Fall back to default if unknown UI type
		handler = h.uiHandlers[UIStoplight]
		h.log.Warn("unknown UI type requested, falling back to stoplight", "requested", uiType)
	}

	handler.GetOpenAPIHTML(w, r)
}

func parseDocsTemplate(path string, logger *slog.Logger) *template.Template {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		logger.Warn("failed to parse custom OpenAPI template, falling back to embedded viewer", "path", path, "error", err)
		return nil
	}
	logger.With("path", path).Info("loaded custom OpenAPI template")
	return tmpl
}

func openAPISpecURL(baseURL string) string {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		return "/info/openapi.json"
	}
	return strings.TrimRight(base, "/") + "/info/openapi.json"
}

// GetAsyncAPIJSON serves the AsyncAPI specification as JSON.
func (h *APIHandler) GetAsyncAPIJSON(w http.ResponseWriter, r *http.Request) {
	h.asyncAPIHandler.GetAsyncAPIJSON(w, r)
}

// GetAsyncAPIHTML serves the AsyncAPI documentation as an interactive HTML page.
func (h *APIHandler) GetAsyncAPIHTML(w http.ResponseWriter, r *http.Request) {
	h.asyncAPIHandler.GetAsyncAPIHTML(w, r)
}
