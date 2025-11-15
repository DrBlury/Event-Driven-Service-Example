package apihandler

import (
	"errors"
	"net/http"

	"drblury/event-driven-service/internal/domain"
)

// SignupNewCustomer accepts signup requests and publishes them as events so downstream
// processors can continue the workflow asynchronously.
func (ah *APIHandler) SignupNewCustomer(w http.ResponseWriter, r *http.Request) {
	if ah == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	signup := &domain.Signup{}
	if ok := ah.ReadRequestBody(w, r, signup); !ok {
		return
	}

	if ah.AppLogic == nil {
		ah.HandleInternalServerError(w, r, errors.New("application logic not configured"), "signup unavailable")
		return
	}

	token := r.Header.Get("Authorization")
	if err := ah.AppLogic.Signup(r.Context(), signup, token); err != nil {
		ah.HandleErrors(w, r, err, "Signup failed")
		return
	}

	if err := ah.AppLogic.EmitSignupEvent(r.Context(), signup); err != nil {
		ah.HandleInternalServerError(w, r, err, "failed to publish signup event")
		return
	}

	ah.RespondWithJSON(w, r, http.StatusCreated, map[string]string{
		"status":  "queued",
		"message": "signup event accepted",
	})
}
