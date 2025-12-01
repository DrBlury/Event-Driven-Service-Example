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

func TestMyWrapperHandlerEnabled(t *testing.T) {
	t.Parallel()

	base := &captureHandler{}
	wrapper := NewMyWrapperHandler(base)

	if !wrapper.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("Enabled should return true")
	}
}

func TestMyWrapperHandlerWithAttrs(t *testing.T) {
	t.Parallel()

	base := &captureHandler{}
	wrapper := NewMyWrapperHandler(base)

	attrs := []slog.Attr{slog.String("key", "value")}
	newHandler := wrapper.WithAttrs(attrs)

	if newHandler == nil {
		t.Fatal("WithAttrs should return non-nil handler")
	}
}

func TestMyWrapperHandlerWithGroup(t *testing.T) {
	t.Parallel()

	base := &captureHandler{}
	wrapper := NewMyWrapperHandler(base)

	newHandler := wrapper.WithGroup("test-group")

	if newHandler == nil {
		t.Fatal("WithGroup should return non-nil handler")
	}
}

func TestNewMyWrapperHandler(t *testing.T) {
	t.Parallel()

	base := &captureHandler{}
	wrapper := NewMyWrapperHandler(base)

	if wrapper == nil {
		t.Fatal("NewMyWrapperHandler should not return nil")
	}
	if wrapper.handler != base {
		t.Error("handler should match")
	}
	if wrapper.buf == nil {
		t.Error("buf should not be nil")
	}
	if wrapper.mutex == nil {
		t.Error("mutex should not be nil")
	}
}

func TestMarshalIndentWrapper(t *testing.T) {
	t.Parallel()

	input := map[string]any{"key": "value"}
	result, err := marshalIndent(input)
	if err != nil {
		t.Fatalf("marshalIndent error: %v", err)
	}
	if len(result) == 0 {
		t.Error("marshalIndent should return non-empty result")
	}
}

func TestMapToSlogAttrsNested(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"nested": map[string]any{
			"inner": "value",
		},
	}
	attrs := mapToSlogAttrs(input, 0, 5)
	if len(attrs) == 0 {
		t.Error("mapToSlogAttrs should return attributes for nested input")
	}
}

func TestMapToSlogAttrsSimple(t *testing.T) {
	t.Parallel()

	input := map[string]any{"key": "value"}
	attrs := mapToSlogAttrs(input, 0, 5)
	if len(attrs) != 1 {
		t.Errorf("expected 1 attribute, got %d", len(attrs))
	}
}

func TestMapToSlogAttrsEmpty(t *testing.T) {
	t.Parallel()

	input := map[string]any{}
	attrs := mapToSlogAttrs(input, 0, 5)
	if len(attrs) != 0 {
		t.Errorf("expected 0 attributes for empty input, got %d", len(attrs))
	}
}

func TestMapToSlogAttrsExceedsMaxDepthMarshalError(t *testing.T) {
	t.Parallel()

	// Create input that would cause marshal error when exceeding depth
	// Use a channel which can't be marshaled to JSON
	input := map[string]any{
		"channel": make(chan int),
	}
	attrs := mapToSlogAttrs(input, 10, 1) // depth > maxDepth
	if len(attrs) != 1 {
		t.Errorf("expected 1 attribute, got %d", len(attrs))
	}
}

func TestMapToSlogAttrsMultipleKeys(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"key1": "value1",
		"key2": 42,
		"key3": true,
		"key4": 3.14,
	}
	attrs := mapToSlogAttrs(input, 0, 5)
	if len(attrs) != 4 {
		t.Errorf("expected 4 attributes, got %d", len(attrs))
	}
}

func TestMapToSlogAttrsDeeplyNested(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"level4": "deep value",
				},
			},
		},
	}
	attrs := mapToSlogAttrs(input, 0, 5)
	if len(attrs) != 1 {
		t.Errorf("expected 1 top-level attribute, got %d", len(attrs))
	}
}

func TestGetJsonAttrBytesSuccess(t *testing.T) {
	t.Parallel()

	base := &captureHandler{}
	wrapper := NewMyWrapperHandler(base)
	rec := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	rec.AddAttrs(slog.String("key", "value"))

	bytes, release, err := wrapper.GetJsonAttrBytes(context.Background(), rec)
	if err != nil {
		t.Fatalf("GetJsonAttrBytes error: %v", err)
	}
	defer release()

	if len(bytes) == 0 {
		t.Error("expected non-empty bytes")
	}
}

func TestMyWrapperHandlerHandleSuccess(t *testing.T) {
	t.Parallel()

	base := &captureHandler{}
	wrapper := NewMyWrapperHandler(base)
	rec := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	rec.AddAttrs(slog.String("simple", "value"))

	if err := wrapper.Handle(context.Background(), rec); err != nil {
		t.Fatalf("Handle error: %v", err)
	}

	if len(base.attrs) == 0 {
		t.Error("expected attributes to be captured")
	}
}

func TestMyWrapperHandlerWithAttrsChained(t *testing.T) {
	t.Parallel()

	base := &captureHandler{}
	wrapper := NewMyWrapperHandler(base)

	// Chain multiple WithAttrs calls
	h1 := wrapper.WithAttrs([]slog.Attr{slog.String("attr1", "value1")})
	h2 := h1.WithAttrs([]slog.Attr{slog.String("attr2", "value2")})

	if h2 == nil {
		t.Fatal("chained WithAttrs should return non-nil")
	}
}

func TestMyWrapperHandlerWithGroupChained(t *testing.T) {
	t.Parallel()

	base := &captureHandler{}
	wrapper := NewMyWrapperHandler(base)

	// Chain multiple WithGroup calls
	h1 := wrapper.WithGroup("group1")
	h2 := h1.WithGroup("group2")

	if h2 == nil {
		t.Fatal("chained WithGroup should return non-nil")
	}
}

func TestMarshalIndentSimpleValue(t *testing.T) {
	t.Parallel()

	input := "simple string"
	result, err := marshalIndent(input)
	if err != nil {
		t.Fatalf("marshalIndent error: %v", err)
	}
	if len(result) == 0 {
		t.Error("marshalIndent should return non-empty result")
	}
}

func TestMarshalIndentArray(t *testing.T) {
	t.Parallel()

	input := []string{"a", "b", "c"}
	result, err := marshalIndent(input)
	if err != nil {
		t.Fatalf("marshalIndent error: %v", err)
	}
	if len(result) == 0 {
		t.Error("marshalIndent should return non-empty result")
	}
}

func TestMarshalIndentNested(t *testing.T) {
	t.Parallel()

	input := map[string]any{
		"outer": map[string]any{
			"inner": "value",
		},
	}
	result, err := marshalIndent(input)
	if err != nil {
		t.Fatalf("marshalIndent error: %v", err)
	}
	if len(result) == 0 {
		t.Error("marshalIndent should return non-empty result")
	}
}

func TestMarshalIndentNilValue(t *testing.T) {
	t.Parallel()

	result, err := marshalIndent(nil)
	if err != nil {
		t.Fatalf("marshalIndent error: %v", err)
	}
	// nil marshals to "null"
	if string(result) != "null" {
		t.Errorf("marshalIndent(nil) = %q, want \"null\"", string(result))
	}
}

func TestMyWrapperHandlerHandleWithNestedMap(t *testing.T) {
	t.Parallel()

	base := &captureHandler{}
	wrapper := NewMyWrapperHandler(base)
	rec := slog.NewRecord(time.Now(), slog.LevelInfo, "nested test", 0)
	rec.AddAttrs(slog.Any("nested", map[string]any{
		"level1": map[string]any{
			"level2": "value",
		},
	}))

	if err := wrapper.Handle(context.Background(), rec); err != nil {
		t.Fatalf("Handle error: %v", err)
	}

	if len(base.attrs) == 0 {
		t.Error("expected attributes to be captured")
	}
}

func TestMyWrapperHandlerWithGroupReturnsMyWrapperHandler(t *testing.T) {
	t.Parallel()

	base := &captureHandler{}
	wrapper := NewMyWrapperHandler(base)

	newHandler := wrapper.WithGroup("test-group")
	_, ok := newHandler.(*MyWrapperHandler)
	if !ok {
		t.Error("WithGroup should return *MyWrapperHandler")
	}
}

func TestMyWrapperHandlerWithAttrsReturnsMyWrapperHandler(t *testing.T) {
	t.Parallel()

	base := &captureHandler{}
	wrapper := NewMyWrapperHandler(base)

	attrs := []slog.Attr{slog.String("key", "value")}
	newHandler := wrapper.WithAttrs(attrs)
	_, ok := newHandler.(*MyWrapperHandler)
	if !ok {
		t.Error("WithAttrs should return *MyWrapperHandler")
	}
}
