package apihandler

import (
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

type APIHandler struct {
	*infohandler.InfoHandler
	AppLogic *usecase.AppLogic
	log      *slog.Logger
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

	options := []infohandler.InfoOption{
		infohandler.WithInfoResponder(resp),
		infohandler.WithBaseURL(baseURL),
		infohandler.WithInfoProvider(infoProvider),
		infohandler.WithSwaggerProvider(swaggerProvider),
		infohandler.WithOpenAPITemplateData(func(_ *http.Request, base string) any {
			return map[string]any{
				"BaseURL": base,
				"SpecURL": openAPISpecURL(base),
			}
		}),
	}

	if tmpl := parseDocsTemplate(docsTemplatePath, logger); tmpl != nil {
		options = append(options, infohandler.WithOpenAPITemplate(tmpl))
	}

	infoHandler := infohandler.NewInfoHandler(options...)

	return &APIHandler{
		InfoHandler: infoHandler,
		AppLogic:    appLogic,
		log:         logger,
	}
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
