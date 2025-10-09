package apihandler

import (
	"drblury/poc-event-signup/internal/domain"
	"drblury/poc-event-signup/internal/usecase"
	"log/slog"
)

//go:generate go tool oapi-codegen -config ./server-std.cfg.yml ./embedded/openapi.json

type APIHandler struct {
	AppLogic *usecase.AppLogic
	Info     *domain.Info
	log      *slog.Logger
	BaseURL  string
}

func NewAPIHandler(appLogic *usecase.AppLogic, info *domain.Info, logger *slog.Logger, baseURL string) *APIHandler {
	return &APIHandler{
		AppLogic: appLogic,
		Info:     info,
		log:      logger,
		BaseURL:  baseURL,
	}
}
