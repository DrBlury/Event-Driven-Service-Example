package apihandler

import (
	"drblury/poc-event-signup/internal/domain"
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

// HandleInternalServerError is a convenient method to log and handle internal server errors.
func (ah *APIHandler) HandleInternalServerError(w http.ResponseWriter, r *http.Request, err error, logMsg ...string) {
	if err == nil {
		err = errors.New("no error information supplied")
	}
	uniqueErrID := uuid.New().String()
	apiError := ErrorPayload{
		ErrorID:   uniqueErrID,
		Code:      500,
		Error:     err.Error(),
		ErrorType: "InternalServerError", // Assuming this is the type string you want
		Timestamp: time.Now().Format(time.RFC3339),
	}

	ah.log.With("error", err.Error()).With("logMessages", logMsg).Error("Internal Server Error")
	ah.RespondWithJSON(w, r, http.StatusInternalServerError, apiError)
}

// HandleBadRequestError is a convenient method to log and handle bad request errors.
func (ah *APIHandler) HandleBadRequestError(w http.ResponseWriter, r *http.Request, err error, logMsg ...string) {
	if err == nil {
		err = errors.New("no error information supplied")
	}
	uniqueErrID := uuid.New().String()
	apiError := ErrorPayload{
		ErrorID:   uniqueErrID,
		Code:      400,
		Error:     err.Error(),
		ErrorType: "BadRequest", // Assuming this is the type string you want
		Timestamp: time.Now().Format(time.RFC3339),
	}

	ah.log.With("error", err.Error()).With("logMessages", logMsg).Warn("Bad Request Error")
	ah.RespondWithJSON(w, r, http.StatusBadRequest, apiError)
}

func (ah *APIHandler) HandleUnauthorizedError(w http.ResponseWriter, r *http.Request, err error, logMsg ...string) {
	if err == nil {
		err = errors.New("no error information supplied")
	}
	uniqueErrID := uuid.New().String()
	apiError := ErrorPayload{
		ErrorID:   uniqueErrID,
		Code:      401,
		Error:     err.Error(),
		ErrorType: "Unauthorized", // Assuming this is the type string you want
		Timestamp: time.Now().Format(time.RFC3339),
	}

	ah.log.With("error", err.Error()).With("logMessages", logMsg).Warn("Unauthorized Error")
	ah.RespondWithJSON(w, r, http.StatusUnauthorized, apiError)
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
	if errors.Is(err, domain.ErrorNotFound) {
		ah.HandleBadRequestError(w, r, err, msgs...)
		return
	}
	if errors.Is(err, domain.ErrorBadRequest) {
		ah.HandleBadRequestError(w, r, err, msgs...)
		return
	}
	if errors.Is(err, domain.ErrorInvalidToken) {
		ah.HandleUnauthorizedError(w, r, err, msgs...)
		return
	}
	if errors.Is(err, domain.ErrorInvalidCredentials) {
		ah.HandleBadRequestError(w, r, err, msgs...)
		return
	}
	if errors.Is(err, domain.ErrorTripicaBusiness) {
		ah.HandleBadRequestError(w, r, err, msgs...)
		return
	}
	if errors.As(err, &domain.ErrTripicaBusiness{}) {
		ah.HandleBadRequestError(w, r, err, msgs...)
		return
	}
	if errors.As(err, &domain.ErrValidations{}) {
		ah.HandleBadRequestError(w, r, err, msgs...)
		return
	}
	if errors.As(err, &domain.ErrSignupFailed{}) {
		ah.HandleBadRequestError(w, r, err, msgs...)
		return
	}
	if errors.Is(err, domain.ErrorUpstreamService) {
		ah.HandleInternalServerError(w, r, err, msgs...)
		return
	}
	ah.HandleInternalServerError(w, r, err)
}
