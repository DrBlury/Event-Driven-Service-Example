package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type ErrorClassifierFunc func(err error) (status int, handled bool)

type statusMeta struct {
	typeLabel string
	logLevel  slog.Level
	logMsg    string
}

type StatusMetadata struct {
	ErrorType string
	LogLevel  slog.Level
	LogMsg    string
}

type ResponderOption func(*Responder)

type Responder struct {
	log             *slog.Logger
	statusMetadata  map[int]statusMeta
	errorClassifier ErrorClassifierFunc
}

func NewResponder(opts ...ResponderOption) *Responder {
	r := &Responder{
		log:            slog.Default(),
		statusMetadata: defaultStatusMetadata(),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(r)
		}
	}
	return r
}

func WithLogger(logger *slog.Logger) ResponderOption {
	return func(r *Responder) {
		if logger != nil {
			r.log = logger
		}
	}
}

func WithErrorClassifier(classifier ErrorClassifierFunc) ResponderOption {
	return func(r *Responder) {
		r.errorClassifier = classifier
	}
}

func WithStatusMetadata(status int, meta StatusMetadata) ResponderOption {
	return func(r *Responder) {
		if r.statusMetadata == nil {
			r.statusMetadata = make(map[int]statusMeta)
		}
		level := meta.LogLevel
		if level == 0 {
			level = slog.LevelError
		}
		msg := meta.LogMsg
		if msg == "" {
			msg = http.StatusText(status)
		}
		r.statusMetadata[status] = statusMeta{
			typeLabel: meta.ErrorType,
			logLevel:  level,
			logMsg:    msg,
		}
	}
}

func defaultStatusMetadata() map[int]statusMeta {
	return map[int]statusMeta{
		http.StatusInternalServerError: {typeLabel: "InternalServerError", logLevel: slog.LevelError, logMsg: "Internal Server Error"},
		http.StatusBadRequest:          {typeLabel: "BadRequest", logLevel: slog.LevelWarn, logMsg: "Bad Request Error"},
		http.StatusUnauthorized:        {typeLabel: "Unauthorized", logLevel: slog.LevelWarn, logMsg: "Unauthorized Error"},
	}
}

func (r *Responder) logger() *slog.Logger {
	if r == nil || r.log == nil {
		return slog.Default()
	}
	return r.log
}

func (r *Responder) Logger() *slog.Logger {
	return r.logger()
}

type ErrorPayload struct {
	ErrorID   string `json:"errorId"`
	Code      int    `json:"code"`
	Error     string `json:"error"`
	ErrorType string `json:"errorType"`
	Timestamp string `json:"timestamp"`
}

func (r *Responder) HandleAPIError(w http.ResponseWriter, req *http.Request, status int, err error, logMsg ...string) {
	if err == nil {
		return
	}

	meta, ok := r.statusMetadata[status]
	if !ok {
		meta = statusMeta{
			typeLabel: "UnknownError",
			logLevel:  slog.LevelError,
			logMsg:    "Unknown Error",
		}
	}

	apiError := ErrorPayload{
		ErrorID:   uuid.New().String(),
		Code:      status,
		Error:     err.Error(),
		ErrorType: meta.typeLabel,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	logger := r.logger().With("error", err.Error()).With("logMessages", logMsg)
	logger.Log(req.Context(), meta.logLevel, meta.logMsg)

	r.RespondWithJSON(w, req, status, apiError)
}

func (r *Responder) HandleInternalServerError(w http.ResponseWriter, req *http.Request, err error, logMsg ...string) {
	r.HandleAPIError(w, req, http.StatusInternalServerError, err, logMsg...)
}

func (r *Responder) HandleBadRequestError(w http.ResponseWriter, req *http.Request, err error, logMsg ...string) {
	r.HandleAPIError(w, req, http.StatusBadRequest, err, logMsg...)
}

func (r *Responder) HandleUnauthorizedError(w http.ResponseWriter, req *http.Request, err error, logMsg ...string) {
	r.HandleAPIError(w, req, http.StatusUnauthorized, err, logMsg...)
}

func (r *Responder) RespondWithJSON(w http.ResponseWriter, req *http.Request, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		r.HandleInternalServerError(w, req, err, "failed to encode response")
	}
}

func (r *Responder) ReadRequestBody(w http.ResponseWriter, req *http.Request, v any) bool {
	if err := json.NewDecoder(req.Body).Decode(v); err != nil {
		r.HandleBadRequestError(w, req, err, "failed to parse request body")
		return false
	}
	return true
}

func (r *Responder) HandleErrors(w http.ResponseWriter, req *http.Request, err error, msgs ...string) {
	if err == nil {
		return
	}

	if r.errorClassifier != nil {
		if status, handled := r.errorClassifier(err); handled {
			r.HandleAPIError(w, req, status, err, msgs...)
			return
		}
	}

	r.HandleInternalServerError(w, req, err, msgs...)
}
