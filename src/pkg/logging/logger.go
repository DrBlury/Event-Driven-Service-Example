package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/samber/lo"
	slogmulti "github.com/samber/slog-multi"
)

// Option configures the logger that is produced by SetLogger.
type Option func(*settings)

type settings struct {
	level        slog.Level
	addSource    bool
	replaceAttr  func([]string, slog.Attr) slog.Attr
	console      consoleSettings
	otel         otelSettings
	setAsDefault bool
	attrs        []slog.Attr
}

type consoleSettings struct {
	enabled bool
	format  Format
	writer  io.Writer
}

type otelSettings struct {
	enabled        bool
	mirrorConsole  bool
	serviceName    string
	serviceVersion string
	endpoint       string
	headers        string
}

func defaultSettings() *settings {
	return &settings{
		level:     slog.LevelInfo,
		addSource: true,
		console: consoleSettings{
			enabled: true,
			format:  FormatText,
			writer:  os.Stdout,
		},
		setAsDefault: true,
	}
}

// SetLogger constructs and installs a slog logger according to the provided options.
// When no options are supplied sensible defaults (text handler @ INFO) are used.
func SetLogger(ctx context.Context, opts ...Option) *slog.Logger {
	if ctx == nil {
		ctx = context.Background()
	}

	cfg := defaultSettings()
	for _, opt := range opts {
		if opt != nil {
			opt(cfg)
		}
	}

	if cfg.console.writer == nil {
		cfg.console.writer = os.Stdout
	}
	if cfg.console.format == "" {
		cfg.console.format = FormatText
	}

	handlers := buildHandlers(ctx, cfg)
	mainHandler := handlers[0]
	if len(handlers) > 1 {
		mainHandler = slogmulti.Fanout(handlers...)
	}

	logger := slog.New(mainHandler)
	if len(cfg.attrs) > 0 {
		args := lo.Map(cfg.attrs, func(attr slog.Attr, _ int) any {
			return attr
		})
		logger = logger.With(args...)
	}

	if cfg.setAsDefault {
		slog.SetDefault(logger)
		slog.SetLogLoggerLevel(cfg.level)
	}

	return logger
}

func buildHandlers(ctx context.Context, cfg *settings) []slog.Handler {
	handlers := make([]slog.Handler, 0, 2)

	if cfg.console.enabled {
		handlers = append(handlers, newConsoleHandler(cfg))
	}

	if cfg.otel.enabled {
		otelHandler := createOtelHandler(ctx, &cfg.otel)
		if otelHandler != nil {
			if cfg.otel.mirrorConsole {
				otelHandler = NewMyWrapperHandler(otelHandler)
			}
			handlers = append(handlers, otelHandler)
		} else if !cfg.console.enabled {
			// Fallback to console output when otel initialisation fails.
			cfg.console.enabled = true
			handlers = append(handlers, newConsoleHandler(cfg))
			fmt.Fprintln(os.Stderr, "logging: OpenTelemetry handler setup failed, falling back to console output")
		}
	}

	if len(handlers) == 0 {
		cfg.console.enabled = true
		handlers = append(handlers, newConsoleHandler(cfg))
	}

	return handlers
}

func newConsoleHandler(cfg *settings) slog.Handler {
	opts := &slog.HandlerOptions{
		Level:       cfg.level,
		AddSource:   cfg.addSource,
		ReplaceAttr: cfg.replaceAttr,
	}

	switch cfg.console.format {
	case FormatJSON:
		return slog.NewJSONHandler(cfg.console.writer, opts)
	case FormatPretty:
		return NewPrettyHandler(opts, cfg.console.writer)
	default:
		return slog.NewTextHandler(cfg.console.writer, opts)
	}
}

// WithConfig applies values from Config to the logger settings.
func WithConfig(cfg *Config) Option {
	return func(s *settings) {
		if cfg == nil {
			return
		}

		if level, err := parseLevel(cfg.Level); err == nil {
			s.level = level
		}

		if cfg.Format != "" {
			if format := ParseFormat(string(cfg.Format)); format != "" {
				s.console.format = format
			}
		}

		s.addSource = cfg.AddSource
		s.console.enabled = cfg.ConsoleEnabled
		s.setAsDefault = cfg.SetAsDefault

		if cfg.OTel.Enabled {
			s.otel.enabled = true
			s.otel.mirrorConsole = cfg.OTel.MirrorToConsole
			s.otel.serviceName = cfg.OTel.ServiceName
			s.otel.serviceVersion = cfg.OTel.ServiceVersion
			s.otel.endpoint = cfg.OTel.Endpoint
			s.otel.headers = cfg.OTel.Headers
			if s.otel.mirrorConsole {
				s.console.enabled = true
			}
		}
	}
}

// WithLevel sets the minimum level for emitted log records.
func WithLevel(level slog.Level) Option {
	return func(s *settings) {
		s.level = level
	}
}

// WithLevelString parses and sets a log level using a string value.
func WithLevelString(level string) Option {
	return func(s *settings) {
		if parsed, err := parseLevel(level); err == nil {
			s.level = parsed
		}
	}
}

// WithJSONFormat enables JSON console output.
func WithJSONFormat() Option {
	return func(s *settings) {
		s.console.enabled = true
		s.console.format = FormatJSON
	}
}

// WithPrettyFormat enables colourful pretty printed console output.
func WithPrettyFormat() Option {
	return func(s *settings) {
		s.console.enabled = true
		s.console.format = FormatPretty
	}
}

// WithTextFormat enables text console output.
func WithTextFormat() Option {
	return func(s *settings) {
		s.console.enabled = true
		s.console.format = FormatText
	}
}

// WithoutConsole disables console logging entirely.
func WithoutConsole() Option {
	return func(s *settings) {
		s.console.enabled = false
	}
}

// WithConsoleWriter overrides the writer used by console handlers.
func WithConsoleWriter(w io.Writer) Option {
	return func(s *settings) {
		s.console.writer = w
	}
}

// WithAddSource toggles inclusion of call-site information.
func WithAddSource(add bool) Option {
	return func(s *settings) {
		s.addSource = add
	}
}

// WithReplaceAttr injects a custom attribute replacer into the handler options.
func WithReplaceAttr(fn func([]string, slog.Attr) slog.Attr) Option {
	return func(s *settings) {
		s.replaceAttr = fn
	}
}

// WithAttrs pre-populates the logger with constant attributes.
func WithAttrs(attrs ...slog.Attr) Option {
	return func(s *settings) {
		s.attrs = append(s.attrs, attrs...)
	}
}

// WithoutGlobal prevents SetLogger from replacing the global slog logger.
func WithoutGlobal() Option {
	return func(s *settings) {
		s.setAsDefault = false
	}
}

// WithOTel enables the OpenTelemetry handler.
func WithOTel(serviceName, serviceVersion string, opts ...OTelOption) Option {
	return func(s *settings) {
		s.otel.enabled = true
		if serviceName != "" {
			s.otel.serviceName = serviceName
		}
		if serviceVersion != "" {
			s.otel.serviceVersion = serviceVersion
		}
		for _, opt := range opts {
			if opt != nil {
				opt(&s.otel)
			}
		}
		if s.otel.mirrorConsole {
			s.console.enabled = true
		}
	}
}

// OTelOption configures OpenTelemetry specific settings.
type OTelOption func(*otelSettings)

// WithOTelEndpoint sets the OTLP endpoint for log export.
func WithOTelEndpoint(endpoint string) OTelOption {
	return func(cfg *otelSettings) {
		cfg.endpoint = endpoint
	}
}

// WithOTelHeaders sets additional OTLP headers.
func WithOTelHeaders(headers string) OTelOption {
	return func(cfg *otelSettings) {
		cfg.headers = headers
	}
}

// WithOTelConsoleMirror ensures console logging stays enabled when OTEL is active.
func WithOTelConsoleMirror() OTelOption {
	return func(cfg *otelSettings) {
		cfg.mirrorConsole = true
	}
}

// ParseFormat normalises user supplied format strings.
func ParseFormat(value string) Format {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "json":
		return FormatJSON
	case "prettyjson", "pretty":
		return FormatPretty
	case "text", "console", "", "plain":
		return FormatText
	default:
		return ""
	}
}

func parseLevel(level string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info", "":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown slog level: %s", level)
	}
}
