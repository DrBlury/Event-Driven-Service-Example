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
