package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ErrorClassifierFunc inspects an error and returns the HTTP status that should
// be used for the response. The boolean indicates whether the error was
// classified and prevents the generic internal server handler from running.
type ErrorClassifierFunc func(err error) (status int, handled bool)

type statusMeta struct {
	typeLabel string
	logLevel  slog.Level
	logMsg    string
}

// StatusMetadata allows callers to customise how particular HTTP status codes
// are logged and represented in error payloads.
type StatusMetadata struct {
	ErrorType string
	LogLevel  slog.Level
	LogMsg    string
}

// ResponderOption follows the functional options pattern used by NewResponder
// to configure optional collaborators.
type ResponderOption func(*Responder)

// Responder centralises error handling, JSON rendering, and logging for HTTP
// handlers. It provides structured error payloads with correlation identifiers
// and consistent log records.
type Responder struct {
	log             *slog.Logger
	statusMetadata  map[int]statusMeta
	errorClassifier ErrorClassifierFunc
}

// NewResponder constructs a Responder with default status metadata and the
// global slog logger. Use ResponderOption functions to override specific
// behaviours.
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

// WithLogger injects a custom slog logger for error reporting and payload
// logging.
func WithLogger(logger *slog.Logger) ResponderOption {
	return func(r *Responder) {
		if logger != nil {
			r.log = logger
		}
	}
}

// WithErrorClassifier installs a classifier used by HandleErrors to derive the
// HTTP status code from returned errors.
func WithErrorClassifier(classifier ErrorClassifierFunc) ResponderOption {
	return func(r *Responder) {
		r.errorClassifier = classifier
	}
}

// WithStatusMetadata overrides the error metadata used for a specific HTTP
// status code.
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

// Logger returns the slog logger used internally by the responder.
func (r *Responder) Logger() *slog.Logger {
	return r.logger()
}

// ErrorPayload is the structured JSON response produced for handled errors.
type ErrorPayload struct {
	ErrorID   string `json:"errorId"`
	Code      int    `json:"code"`
	Error     string `json:"error"`
	ErrorType string `json:"errorType"`
	Timestamp string `json:"timestamp"`
}

// HandleAPIError renders a structured JSON response for the supplied HTTP
// status and logs the payload using the configured logger.
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

// HandleInternalServerError is a shortcut that reports a 500 status code.
func (r *Responder) HandleInternalServerError(w http.ResponseWriter, req *http.Request, err error, logMsg ...string) {
	r.HandleAPIError(w, req, http.StatusInternalServerError, err, logMsg...)
}

// HandleBadRequestError reports client validation errors using HTTP 400.
func (r *Responder) HandleBadRequestError(w http.ResponseWriter, req *http.Request, err error, logMsg ...string) {
	r.HandleAPIError(w, req, http.StatusBadRequest, err, logMsg...)
}

// HandleUnauthorizedError reports authentication failures using HTTP 401.
func (r *Responder) HandleUnauthorizedError(w http.ResponseWriter, req *http.Request, err error, logMsg ...string) {
	r.HandleAPIError(w, req, http.StatusUnauthorized, err, logMsg...)
}

// RespondWithJSON serialises the provided value and writes it to the response
// using the supplied status code.
func (r *Responder) RespondWithJSON(w http.ResponseWriter, req *http.Request, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		r.HandleInternalServerError(w, req, err, "failed to encode response")
	}
}

// ReadRequestBody parses the request body into the provided value and handles
// malformed content by returning a JSON error response.
func (r *Responder) ReadRequestBody(w http.ResponseWriter, req *http.Request, v any) bool {
	if err := json.NewDecoder(req.Body).Decode(v); err != nil {
		r.HandleBadRequestError(w, req, err, "failed to parse request body")
		return false
	}
	return true
}

// HandleErrors inspects the supplied error using the configured classifier and
// emits an appropriate JSON response.
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
