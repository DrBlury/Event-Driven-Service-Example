package apihandler

import (
	"errors"
	"net/http"

	"drblury/event-driven-service/internal/domain"
	generator "drblury/event-driven-service/internal/server/gen"
	"drblury/event-driven-service/internal/usecase"
	infohandler "drblury/event-driven-service/pkg/api/info"
	"drblury/event-driven-service/pkg/api/responder"
	"log/slog"
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

	infoHandler := infohandler.NewInfoHandler(
		infohandler.WithInfoResponder(resp),
		infohandler.WithBaseURL(baseURL),
		infohandler.WithInfoProvider(infoProvider),
		infohandler.WithSwaggerProvider(swaggerProvider),
	)

	return &APIHandler{
		InfoHandler: infoHandler,
		AppLogic:    appLogic,
		log:         logger,
	}
}
