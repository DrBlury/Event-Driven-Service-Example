package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// ErrorClassifierFunc inspects an error and returns the HTTP status that should
// be used for the response. The boolean indicates whether the error was
// classified and prevents the generic internal server handler from running.
type ErrorClassifierFunc func(err error) (status int, handled bool)

type statusMeta struct {
	typeURI  string
	title    string
	logLevel slog.Level
	logMsg   string
}

// StatusMetadata allows callers to customise how particular HTTP status codes
// are logged and represented in error payloads.
type StatusMetadata struct {
	TypeURI  string
	Title    string
	LogLevel slog.Level
	LogMsg   string
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
		title := meta.Title
		if title == "" {
			title = http.StatusText(status)
		}
		msg := meta.LogMsg
		if msg == "" {
			msg = title
		}
		r.statusMetadata[status] = statusMeta{
			typeURI:  meta.TypeURI,
			title:    title,
			logLevel: level,
			logMsg:   msg,
		}
	}
}

func defaultStatusMetadata() map[int]statusMeta {
	return map[int]statusMeta{
		http.StatusInternalServerError: {title: http.StatusText(http.StatusInternalServerError), logLevel: slog.LevelError, logMsg: "Internal Server Error"},
		http.StatusBadRequest:          {title: http.StatusText(http.StatusBadRequest), logLevel: slog.LevelWarn, logMsg: "Bad Request"},
		http.StatusUnauthorized:        {title: http.StatusText(http.StatusUnauthorized), logLevel: slog.LevelWarn, logMsg: "Unauthorized"},
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

// ProblemDetails aligns HTTP error responses with RFC 9457 problem documents.
type ProblemDetails struct {
	Type      string `json:"type,omitempty"`
	Title     string `json:"title"`
	Status    int    `json:"status"`
	Detail    string `json:"detail,omitempty"`
	Instance  string `json:"instance,omitempty"`
	TraceID   string `json:"traceId,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// HandleAPIError renders a structured JSON response for the supplied HTTP
// status and logs the payload using the configured logger.
func (r *Responder) HandleAPIError(w http.ResponseWriter, req *http.Request, status int, err error, logMsg ...string) {
	if err == nil {
		return
	}

	meta, ok := r.statusMetadata[status]
	if !ok {
		meta = statusMeta{}
	}

	if meta.logLevel == 0 {
		meta.logLevel = slog.LevelError
	}

	if meta.title == "" {
		meta.title = http.StatusText(status)
	}

	if meta.logMsg == "" {
		meta.logMsg = meta.title
	}

	traceID := CreateULID()
	problem := ProblemDetails{
		Type:      meta.typeURI,
		Title:     meta.title,
		Status:    status,
		Detail:    err.Error(),
		TraceID:   traceID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if problem.Type == "" {
		problem.Type = fmt.Sprintf("https://httpstatuses.io/%d", status)
	}

	if req != nil && req.URL != nil {
		problem.Instance = req.URL.RequestURI()
	}

	logger := r.logger().With(
		"error", err.Error(),
		"traceId", traceID,
		"status", status,
		"logMessages", logMsg,
	)
	logger.Log(req.Context(), meta.logLevel, meta.logMsg)

	r.respondWithJSON(w, req, status, problem, "application/problem+json")
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
	r.respondWithJSON(w, req, status, v, "application/json")
}

func (r *Responder) respondWithJSON(w http.ResponseWriter, req *http.Request, status int, payload any, contentType string) {
	if w == nil {
		return
	}

	if contentType == "" {
		contentType = "application/json"
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(payload); err != nil {
		r.logger().Error("failed to encode response", "error", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	if _, err := w.Write(buf.Bytes()); err != nil {
		r.logger().Error("failed to write response", "error", err)
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
