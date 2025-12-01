package logging

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"
)

type stubSlogHandler struct {
	handleFunc func(context.Context, slog.Record) error
}

func (s stubSlogHandler) Enabled(context.Context, slog.Level) bool { return true }
func (s stubSlogHandler) Handle(ctx context.Context, r slog.Record) error {
	if s.handleFunc != nil {
		return s.handleFunc(ctx, r)
	}
	return nil
}
func (s stubSlogHandler) WithAttrs([]slog.Attr) slog.Handler { return s }
func (s stubSlogHandler) WithGroup(string) slog.Handler      { return s }

type failingWriter struct{}

func (f failingWriter) Write([]byte) (int, error) { return 0, errors.New("write") }

func TestNewPrettyHandlerHandlesLevels(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	h := NewPrettyHandler(&slog.HandlerOptions{Level: slog.LevelDebug}, buf)

	levels := []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}
	for _, lvl := range levels {
		rec := slog.NewRecord(time.Now(), lvl, "message", 0)
		rec.AddAttrs(slog.String("key", "value"))
		if err := h.Handle(context.Background(), rec); err != nil {
			t.Fatalf("unexpected handle error: %v", err)
		}
	}

	if buf.Len() == 0 {
		t.Fatal("expected output to be written")
	}

	if !h.Enabled(context.Background(), slog.LevelDebug) {
		t.Fatal("expected handler to be enabled")
	}
}

func TestPrettyHandlerWithAttrsAndGroups(t *testing.T) {
	t.Parallel()

	h := NewPrettyHandler(nil, &bytes.Buffer{})
	hAttr := h.WithAttrs([]slog.Attr{slog.String("k", "v")}).(*PrettyJsonHandler)
	hGroup := h.WithGroup("grp").(*PrettyJsonHandler)

	if hAttr.writer != h.writer || hGroup.writer != h.writer {
		t.Fatal("expected writer to be shared across derived handlers")
	}
}

func TestPrettyHandlerSuppressDefaults(t *testing.T) {
	t.Parallel()

	replacer := suppressDefaults(nil)
	if got := replacer(nil, slog.Time("time", time.Now())); got.Key != "" {
		t.Fatal("expected default attributes to be suppressed")
	}
	if got := replacer(nil, slog.String("custom", "value")); got.Key != "custom" {
		t.Fatal("expected custom attribute passthrough")
	}
}

func TestPrettyHandlerComputeAttrsErrors(t *testing.T) {
	t.Parallel()

	h := NewPrettyHandler(nil, &bytes.Buffer{})
	h.slogHandler = stubSlogHandler{handleFunc: func(context.Context, slog.Record) error {
		return errors.New("fail")
	}}
	if _, err := h.computeAttrs(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)); err == nil {
		t.Fatal("expected computeAttrs failure when inner handler fails")
	}

	invalid := &PrettyJsonHandler{
		buf:    &bytes.Buffer{},
		mutex:  &sync.Mutex{},
		writer: &bytes.Buffer{},
	}
	invalid.slogHandler = stubSlogHandler{handleFunc: func(ctx context.Context, r slog.Record) error {
		_, _ = invalid.buf.WriteString("not-json")
		return nil
	}}
	if _, err := invalid.computeAttrs(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)); err == nil {
		t.Fatal("expected error when JSON decoding fails")
	}
}

func TestPrettyHandlerHandleWriteError(t *testing.T) {
	t.Parallel()

	h := NewPrettyHandler(nil, failingWriter{})
	if err := h.Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)); err == nil {
		t.Fatal("expected write error to propagate")
	}
}

func TestColorize(t *testing.T) {
	t.Parallel()

	res := colorize(31, "msg")
	if !strings.Contains(res, "msg") {
		t.Fatal("expected message to be wrapped by colorizer")
	}
}

func TestNewPrettyHandlerNilOpts(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	h := NewPrettyHandler(nil, buf)

	if h == nil {
		t.Fatal("NewPrettyHandler should not return nil")
	}
	if !h.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("handler should be enabled for info level with default options")
	}
}

func TestNewPrettyHandlerNoWriter(t *testing.T) {
	t.Parallel()

	h := NewPrettyHandler(nil)

	if h == nil {
		t.Fatal("NewPrettyHandler should not return nil")
	}
	// Should default to os.Stdout
	if h.writer == nil {
		t.Error("writer should not be nil")
	}
}

func TestNewPrettyHandlerNilWriter(t *testing.T) {
	t.Parallel()

	h := NewPrettyHandler(nil, nil)

	if h == nil {
		t.Fatal("NewPrettyHandler should not return nil")
	}
	// When nil is explicitly passed, should default to os.Stdout
	if h.writer == nil {
		t.Error("writer should default to os.Stdout when nil")
	}
}

func TestPrettyHandlerAllLogLevels(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	h := NewPrettyHandler(&slog.HandlerOptions{Level: slog.LevelDebug}, buf)

	levels := []slog.Level{
		slog.LevelDebug,
		slog.LevelInfo,
		slog.LevelWarn,
		slog.LevelError,
	}

	for _, level := range levels {
		buf.Reset()
		rec := slog.NewRecord(time.Now(), level, "test message", 0)
		rec.AddAttrs(slog.String("key", "value"))

		if err := h.Handle(context.Background(), rec); err != nil {
			t.Errorf("Handle failed for level %v: %v", level, err)
		}

		if buf.Len() == 0 {
			t.Errorf("Expected output for level %v", level)
		}
	}
}

func TestSuppressDefaultsWithNext(t *testing.T) {
	t.Parallel()

	customReplacer := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == "custom" {
			return slog.String("custom", "replaced")
		}
		return a
	}

	replacer := suppressDefaults(customReplacer)

	// Test that default keys are suppressed
	if got := replacer(nil, slog.String(slog.TimeKey, "2024-01-01")); got.Key != "" {
		t.Error("TimeKey should be suppressed")
	}
	if got := replacer(nil, slog.String(slog.LevelKey, "INFO")); got.Key != "" {
		t.Error("LevelKey should be suppressed")
	}
	if got := replacer(nil, slog.String(slog.MessageKey, "msg")); got.Key != "" {
		t.Error("MessageKey should be suppressed")
	}

	// Test that custom replacer is called for non-default keys
	if got := replacer(nil, slog.String("custom", "value")); got.Value.String() != "replaced" {
		t.Error("custom replacer should be called")
	}

	// Test passthrough for other keys
	if got := replacer(nil, slog.String("other", "value")); got.Key != "other" {
		t.Error("other keys should pass through")
	}
}

func TestPrettyHandlerWithComplexAttrs(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	h := NewPrettyHandler(&slog.HandlerOptions{Level: slog.LevelDebug}, buf)

	rec := slog.NewRecord(time.Now(), slog.LevelInfo, "complex message", 0)
	rec.AddAttrs(
		slog.String("string", "value"),
		slog.Int("int", 42),
		slog.Bool("bool", true),
		slog.Float64("float", 3.14),
		slog.Group("group",
			slog.String("nested", "value"),
		),
	)

	if err := h.Handle(context.Background(), rec); err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("Expected output for complex attributes")
	}
}

func TestColorizeVariants(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		colorCode int
		value     string
	}{
		{black, "black"},
		{red, "red"},
		{green, "green"},
		{yellow, "yellow"},
		{blue, "blue"},
		{magenta, "magenta"},
		{cyan, "cyan"},
		{lightGray, "lightGray"},
		{darkGray, "darkGray"},
		{lightRed, "lightRed"},
		{lightGreen, "lightGreen"},
		{lightYellow, "lightYellow"},
		{lightBlue, "lightBlue"},
		{lightMagenta, "lightMagenta"},
		{lightCyan, "lightCyan"},
		{white, "white"},
	}

	for _, tc := range testCases {
		result := colorize(tc.colorCode, tc.value)
		if !strings.Contains(result, tc.value) {
			t.Errorf("colorize(%d, %q) should contain %q", tc.colorCode, tc.value, tc.value)
		}
		if !strings.Contains(result, reset) {
			t.Errorf("colorize result should contain reset sequence")
		}
	}
}
