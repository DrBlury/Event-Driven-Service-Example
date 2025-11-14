package logging

import (
	"fmt"
	"log/slog"
)

// RestyLogger adapts slog.Logger to the logger interface expected by the Resty
// HTTP client.
type RestyLogger struct {
	logger *slog.Logger
}

// NewRestyLogger wraps the supplied slog.Logger in a Resty compatible adapter.
func NewRestyLogger(logger *slog.Logger) *RestyLogger {
	return &RestyLogger{
		logger: logger,
	}
}

// Errorf logs errors using slog at error level.
func (r *RestyLogger) Errorf(format string, v ...any) {
	r.logger.Error(fmt.Sprintf(format, v...))
}

// Warnf logs formatted messages at warn level.
func (r *RestyLogger) Warnf(format string, v ...any) {
	r.logger.Warn(fmt.Sprintf(format, v...))
}

// Infof logs formatted messages at info level.
func (r *RestyLogger) Infof(format string, v ...any) {
	r.logger.Info(fmt.Sprintf(format, v...))
}

// Debugf logs formatted messages at debug level.
func (r *RestyLogger) Debugf(format string, v ...any) {
	r.logger.Debug(fmt.Sprintf(format, v...))
}
