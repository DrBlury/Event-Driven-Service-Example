package apihandler

import (
	"drblury/event-driven-service/internal/domain"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ErrorPayload defines the structure of the error response payload.
type ErrorPayload struct {
	ErrorID   string `json:"errorId"`
	Code      int    `json:"code"`
	Error     string `json:"error"`
	ErrorType string `json:"errorType"`
	Timestamp string `json:"timestamp"`
}

var errorTypeMap = map[int]struct {
	Type      string
	LogMethod string
	LogMsg    string
}{
	http.StatusInternalServerError: {Type: "InternalServerError", LogMethod: "Error", LogMsg: "Internal Server Error"},
	http.StatusBadRequest:          {Type: "BadRequest", LogMethod: "Warn", LogMsg: "Bad Request Error"},
	http.StatusUnauthorized:        {Type: "Unauthorized", LogMethod: "Warn", LogMsg: "Unauthorized Error"},
}

func (ah *APIHandler) HandleAPIError(w http.ResponseWriter, r *http.Request, status int, err error, logMsg ...string) {
	if err == nil {
		return
	}
	meta, ok := errorTypeMap[status]
	if !ok {
		meta = struct {
			Type      string
			LogMethod string
			LogMsg    string
		}{"UnknownError", "Error", "Unknown Error"}
	}
	uniqueErrID := uuid.New().String()
	apiError := ErrorPayload{
		ErrorID:   uniqueErrID,
		Code:      status,
		Error:     err.Error(),
		ErrorType: meta.Type,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	logger := ah.log.With("error", err.Error()).With("logMessages", logMsg)
	switch meta.LogMethod {
	case "Warn":
		logger.Warn(meta.LogMsg)
	default:
		logger.Error(meta.LogMsg)
	}
	ah.RespondWithJSON(w, r, status, apiError)
}

// Convenience wrappers for compatibility
func (ah *APIHandler) HandleInternalServerError(w http.ResponseWriter, r *http.Request, err error, logMsg ...string) {
	ah.HandleAPIError(w, r, http.StatusInternalServerError, err, logMsg...)
}
func (ah *APIHandler) HandleBadRequestError(w http.ResponseWriter, r *http.Request, err error, logMsg ...string) {
	ah.HandleAPIError(w, r, http.StatusBadRequest, err, logMsg...)
}
func (ah *APIHandler) HandleUnauthorizedError(w http.ResponseWriter, r *http.Request, err error, logMsg ...string) {
	ah.HandleAPIError(w, r, http.StatusUnauthorized, err, logMsg...)
}

func (ah *APIHandler) RespondWithJSON(w http.ResponseWriter, r *http.Request, status int, v any) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		ah.HandleInternalServerError(w, r, err, "Failed to encode response")
	}
}

func (ah *APIHandler) ReadRequestBody(w http.ResponseWriter, r *http.Request, v any) bool {
	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		ah.HandleBadRequestError(w, r, err, "Failed to parse request body")
	}
	return true
}

func (ah *APIHandler) HandleErrors(w http.ResponseWriter, r *http.Request, err error, msgs ...string) {
	if err == nil {
		return // Do nothing if error is nil
	}
	switch {
	case errors.Is(err, domain.ErrorUpstreamService):
		ah.HandleInternalServerError(w, r, err, msgs...)
	case errors.Is(err, domain.ErrorNotFound),
		errors.Is(err, domain.ErrorBadRequest),
		errors.As(err, &domain.ErrValidations{}):
		ah.HandleBadRequestError(w, r, err, msgs...)
	default:
		ah.HandleInternalServerError(w, r, err)
	}
}
