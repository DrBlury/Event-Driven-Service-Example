package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"sync"
)

const (
	reset = "\033[0m"

	black        = 30
	red          = 31
	green        = 32
	yellow       = 33
	blue         = 34
	magenta      = 35
	cyan         = 36
	lightGray    = 37
	darkGray     = 90
	lightRed     = 91
	lightGreen   = 92
	lightYellow  = 93
	lightBlue    = 94
	lightMagenta = 95
	lightCyan    = 96
	white        = 97
)

func NewPrettyHandler(opts *slog.HandlerOptions, writers ...io.Writer) *PrettyJsonHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	var writer io.Writer = os.Stdout
	if len(writers) > 0 && writers[0] != nil {
		writer = writers[0]
	}
	b := &bytes.Buffer{}
	return &PrettyJsonHandler{
		buf: b,
		slogHandler: slog.NewJSONHandler(b, &slog.HandlerOptions{
			Level:       opts.Level,
			AddSource:   opts.AddSource,
			ReplaceAttr: suppressDefaults(opts.ReplaceAttr),
		}),
		writer: writer,
		mutex:  &sync.Mutex{},
	}
}

func colorize(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, reset)
}

type PrettyJsonHandler struct {
	slogHandler slog.Handler
	buf         *bytes.Buffer
	mutex       *sync.Mutex
	writer      io.Writer
}

func (h *PrettyJsonHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.slogHandler.Enabled(ctx, level)
}

func (h *PrettyJsonHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &PrettyJsonHandler{
		slogHandler: h.slogHandler.WithAttrs(attrs),
		buf:         h.buf,
		mutex:       h.mutex,
		writer:      h.writer,
	}
}

func (h *PrettyJsonHandler) WithGroup(name string) slog.Handler {
	return &PrettyJsonHandler{
		slogHandler: h.slogHandler.WithGroup(name),
		buf:         h.buf,
		mutex:       h.mutex,
		writer:      h.writer,
	}
}

const (
	timeFormat = "[15:04:05.000]"
)

func (h *PrettyJsonHandler) Handle(ctx context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = colorize(darkGray, level)
	case slog.LevelInfo:
		level = colorize(cyan, level)
	case slog.LevelWarn:
		level = colorize(lightYellow, level)
	case slog.LevelError:
		level = colorize(lightRed, level)
	}

	attrs, err := h.computeAttrs(ctx, r)
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(attrs, "", "  ")
	if err != nil {
		return fmt.Errorf("error when marshaling attrs: %w", err)
	}

	_, err = fmt.Fprintln(
		h.writer,
		colorize(lightGray, r.Time.Format(timeFormat)),
		level,
		colorize(white, r.Message),
		colorize(darkGray, string(bytes)),
	)
	if err != nil {
		return fmt.Errorf("failed to write pretty log: %w", err)
	}

	return nil
}

func suppressDefaults(
	next func([]string, slog.Attr) slog.Attr,
) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey ||
			a.Key == slog.LevelKey ||
			a.Key == slog.MessageKey {
			return slog.Attr{}
		}
		if next == nil {
			return a
		}
		return next(groups, a)
	}
}

func (h *PrettyJsonHandler) computeAttrs(
	ctx context.Context,
	r slog.Record,
) (map[string]any, error) {
	h.mutex.Lock()
	defer func() {
		h.buf.Reset()
		h.mutex.Unlock()
	}()
	if err := h.slogHandler.Handle(ctx, r); err != nil {
		return nil, fmt.Errorf("error when calling inner handler's Handle: %w", err)
	}

	var attrs map[string]any
	err := json.Unmarshal(h.buf.Bytes(), &attrs)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshaling inner handler's Handle result: %w", err)
	}
	return attrs, nil
}
