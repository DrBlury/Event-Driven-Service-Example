package apihandler

import (
	"errors"
	"net/http"

	"drblury/event-driven-service/internal/domain"
)

// CreateExampleRecord accepts example data and publishes it as a proto event.
func (ah *APIHandler) CreateExampleRecord(w http.ResponseWriter, r *http.Request) {
	if ah == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	record := &domain.ExampleRecord{}
	if ok := ah.ReadRequestBody(w, r, record); !ok {
		return
	}

	if ah.AppLogic == nil {
		ah.HandleInternalServerError(w, r, errors.New("application logic not configured"), "example processing unavailable")
		return
	}

	token := r.Header.Get("Authorization")
	if err := ah.AppLogic.HandleExample(r.Context(), record, token); err != nil {
		ah.HandleErrors(w, r, err, "Example processing failed")
		return
	}

	ah.RespondWithJSON(w, r, http.StatusCreated, map[string]string{
		"status":  "queued",
		"message": "example event accepted",
	})
}
