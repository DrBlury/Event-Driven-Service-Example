package logging

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"
)

type captureHandler struct {
	attrs []slog.Attr
}

func (c *captureHandler) Enabled(context.Context, slog.Level) bool { return true }

func (c *captureHandler) Handle(_ context.Context, r slog.Record) error {
	c.attrs = nil
	r.Attrs(func(a slog.Attr) bool {
		c.attrs = append(c.attrs, a)
		return true
	})
	return nil
}

func (c *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return c }
func (c *captureHandler) WithGroup(string) slog.Handler      { return c }

type failingJSONHandler struct{ err error }

func (f failingJSONHandler) Enabled(context.Context, slog.Level) bool  { return true }
func (f failingJSONHandler) Handle(context.Context, slog.Record) error { return f.err }
func (f failingJSONHandler) WithAttrs([]slog.Attr) slog.Handler        { return f }
func (f failingJSONHandler) WithGroup(string) slog.Handler             { return f }

func TestMyWrapperHandlerHandleFlattensAttrs(t *testing.T) {
	t.Parallel()

	base := &captureHandler{}
	wrapper := NewMyWrapperHandler(base)
	rec := slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)
	rec.AddAttrs(slog.Any("payload", map[string]any{"key": "value"}))

	if err := wrapper.Handle(context.Background(), rec); err != nil {
		t.Fatalf("unexpected handle error: %v", err)
	}

	if len(base.attrs) == 0 {
		t.Fatal("expected attributes to be forwarded to wrapped handler")
	}
}

func TestMyWrapperHandlerGetJsonAttrBytesError(t *testing.T) {
	t.Parallel()

	h := &MyWrapperHandler{
		handler:     &captureHandler{},
		jsonHandler: failingJSONHandler{err: errors.New("boom")},
		buf:         &bytes.Buffer{},
		mutex:       &sync.Mutex{},
	}

	_, _, err := h.GetJsonAttrBytes(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0))
	if err == nil {
		t.Fatal("expected error from json handler to propagate")
	}
}

func TestMyWrapperHandlerHandleInvalidJSON(t *testing.T) {
	t.Parallel()

	h := &MyWrapperHandler{
		handler: &captureHandler{},
		buf:     &bytes.Buffer{},
		mutex:   &sync.Mutex{},
	}
	h.jsonHandler = stubSlogHandler{handleFunc: func(context.Context, slog.Record) error {
		_, _ = h.buf.WriteString("invalid")
		return nil
	}}

	if err := h.Handle(context.Background(), slog.NewRecord(time.Now(), slog.LevelInfo, "msg", 0)); err == nil {
		t.Fatal("expected error when JSON cannot be unmarshalled")
	}
}

func TestMapToSlogAttrsDepthLimit(t *testing.T) {
	t.Parallel()

	input := map[string]any{"deep": map[string]any{"value": "end"}}
	attrs := mapToSlogAttrs(input, 2, 1)
	if len(attrs) != 1 {
		t.Fatalf("expected single attribute, got %d", len(attrs))
	}
	if attrs[0].Value.Kind() != slog.KindString {
		t.Fatalf("expected string representation beyond max depth, got %v", attrs[0].Value.Kind())
	}
}
