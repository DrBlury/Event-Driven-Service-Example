package apihandler

import (
	"net/http"

	"drblury/event-driven-service/internal/domain"
	"drblury/event-driven-service/internal/observability"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("drblury/event-driven-service/internal/server/handler/apihandler")

// CreateExampleRecord accepts example data and publishes it as a proto event.
func (ah *APIHandler) CreateExampleRecord(w http.ResponseWriter, r *http.Request) {
	if ah == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	ctx, span := tracer.Start(r.Context(), "http.create_example_record")
	defer span.End()

	record := &domain.ExampleRecord{}
	if ok := ah.ReadRequestBody(w, r, record); !ok {
		return
	}

	span.SetAttributes(attribute.String("example.record_id", record.GetRecordId()))
	log := observability.Logger(ctx, ah.log).With(
		"record_id", record.GetRecordId(),
		"path", r.URL.Path,
		"method", r.Method,
	)

	if ah.AppLogic == nil {
		err := observability.Builder(ctx, "http.create_example_record", "app_logic_missing").
			Public("example processing is unavailable").
			New("application logic not configured")
		observability.RecordError(span, err)
		log.Error("application logic not configured for example request", "error", err)
		ah.HandleInternalServerError(w, r.WithContext(ctx), err, "example processing unavailable")
		return
	}

	token := r.Header.Get("Authorization")
	if err := ah.AppLogic.HandleExample(ctx, record, token); err != nil {
		wrapped := observability.Builder(ctx, "http.create_example_record", "handle_example_failed").
			Public("example processing failed").
			With("record_id", record.GetRecordId(), "path", r.URL.Path).
			Wrap(err)
		observability.RecordError(span, wrapped)
		log.Error("failed to process example record", "error", wrapped)
		ah.HandleErrors(w, r.WithContext(ctx), wrapped, "Example processing failed")
		return
	}

	log.Info("accepted example record for processing")

	ah.RespondWithJSON(w, r.WithContext(ctx), http.StatusCreated, map[string]string{
		"status":  "queued",
		"message": "example event accepted",
	})
}
