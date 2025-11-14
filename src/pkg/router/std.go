package router

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	oapiMW "github.com/oapi-codegen/nethttp-middleware"
)

// New returns a new *http.ServeMux to be used for HTTP request routing.
func New(apiHandle http.Handler, cfg *Config, logger *slog.Logger, swagger *openapi3.T) *http.ServeMux {
	// Create a new ServeMux
	mux := http.NewServeMux()
	handler := addMiddlewares(apiHandle, cfg, logger, swagger)
	mux.Handle("/", handler)
	return mux
}

// addMiddleware chains together all the middleware functions.
func addMiddlewares(handler http.Handler, cfg *Config, logger *slog.Logger, swagger *openapi3.T) http.Handler {
	// The first handler is the last one to be called
	handler = oapiMiddleware(handler, swagger)
	handler = corsMiddleware(handler, cfg.CORS)
	handler = timeoutMiddleware(handler, cfg.Timeout)
	logger.With("QuietdownRoutes", cfg.QuietdownRoutes, "HideHeaders", cfg.HideHeaders).Debug("Config for logging middleware")
	handler = loggingMiddleware(handler, logger, cfg.QuietdownRoutes, cfg.HideHeaders)
	return handler
}

func oapiMiddleware(handler http.Handler, swagger *openapi3.T) http.Handler {
	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	// Validate requests against OpenAPI spec
	validatorOptions := &oapiMW.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: func(c context.Context, input *openapi3filter.AuthenticationInput) error {
				return nil
			},
		},
	}

	return oapiMW.OapiRequestValidatorWithOptions(swagger, validatorOptions)(handler)
}

func loggingMiddleware(next http.Handler, logger *slog.Logger, quietdownRoutes []string, hideHeaders []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !shouldQuietRoute(r.URL.Path, quietdownRoutes) {
			headers := cloneHeaders(r.Header)
			redactHeaders(headers, hideHeaders)

			attrs := []any{
				"Path", r.URL.Path,
				"Method", r.Method,
				"Header", headers,
			}

			if r.ContentLength > 0 {
				attrs = append(attrs, "ContentLength", r.ContentLength)
			}

			logger.With(attrs...).Debug("Request")
		}

		next.ServeHTTP(w, r)
	})
}

// corsMiddleware adds CORS headers based on the provided configuration.
func corsMiddleware(next http.Handler, cfg CORSConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		// Check if the origin is allowed
		allowedOrigin := false
		for _, o := range cfg.Origins {
			if o == "*" || o == origin {
				allowedOrigin = true
				break
			}
		}

		if allowedOrigin {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}

		// Handle other CORS headers
		if r.Method == http.MethodOptions {
			// Preflight request handling
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.Methods, ","))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.Headers, ","))
			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// timeoutMiddleware adds timeout handling to requests.
func timeoutMiddleware(next http.Handler, timeout time.Duration) http.Handler {
	return http.TimeoutHandler(next, timeout, "Timeout")
}

func shouldQuietRoute(path string, quietdownRoutes []string) bool {
	for _, quietPath := range quietdownRoutes {
		if path == quietPath {
			return true
		}
	}

	return false
}

func cloneHeaders(src http.Header) http.Header {
	headers := make(http.Header, len(src))
	for k, v := range src {
		copied := make([]string, len(v))
		copy(copied, v)
		headers[k] = copied
	}

	return headers
}

func redactHeaders(headers http.Header, hideHeaders []string) {
	for _, header := range hideHeaders {
		canonical := http.CanonicalHeaderKey(header)
		values, exists := headers[canonical]
		if !exists {
			continue
		}

		redactedLen := 0
		for _, value := range values {
			redactedLen += len(value)
		}

		headers[canonical] = []string{fmt.Sprintf("[REDACTED - %d bytes]", redactedLen)}
	}
}
