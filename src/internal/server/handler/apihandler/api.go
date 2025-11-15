package apihandler

import (
	"errors"
	"net/http"

	"drblury/event-driven-service/internal/domain"
	generatedAPI "drblury/event-driven-service/internal/server/_gen"
	"drblury/event-driven-service/internal/usecase"
	"drblury/event-driven-service/pkg/api"
	"log/slog"
)

type APIHandler struct {
	*api.InfoHandler
	AppLogic *usecase.AppLogic
	log      *slog.Logger
}

func NewAPIHandler(
	appLogic *usecase.AppLogic,
	info *domain.Info,
	logger *slog.Logger,
	baseURL string,
) *APIHandler {
	responder := api.NewResponder(
		api.WithLogger(logger),
		api.WithErrorClassifier(func(err error) (int, bool) {
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
		swagger, err := generatedAPI.GetSwagger()
		if err != nil {
			return nil, err
		}
		return swagger.MarshalJSON()
	}

	infoProvider := func() any {
		return info
	}

	infoHandler := api.NewInfoHandler(
		api.WithInfoResponder(responder),
		api.WithBaseURL(baseURL),
		api.WithInfoProvider(infoProvider),
		api.WithSwaggerProvider(swaggerProvider),
	)

	return &APIHandler{
		InfoHandler: infoHandler,
		AppLogic:    appLogic,
		log:         logger,
	}
}
