package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/samber/lo"
)

// MyWrapperHandler wraps another slog.Handler and flattens nested JSON payloads
// into slog attributes so they are easier to filter in downstream exporters.
type MyWrapperHandler struct {
	handler     slog.Handler
	jsonHandler slog.Handler
	buf         *bytes.Buffer
	mutex       *sync.Mutex
}

// NewMyWrapperHandler creates a new MyWrapperHandler.
func NewMyWrapperHandler(handler slog.Handler) *MyWrapperHandler {
	buf := &bytes.Buffer{} // Create a buffer pointer
	return &MyWrapperHandler{
		handler: handler,
		jsonHandler: slog.NewJSONHandler(buf, &slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}),
		buf:   buf,
		mutex: &sync.Mutex{},
	}
}

// Enabled delegates the Enabled check to the wrapped handler.
func (h *MyWrapperHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// GetJsonAttrBytes renders the wrapped handler output as JSON and returns the
// raw bytes so they can be transformed into slog attributes.
func (h *MyWrapperHandler) GetJsonAttrBytes(
	ctx context.Context,
	r slog.Record,
) ([]byte, func(), error) {
	h.mutex.Lock()
	defer func() {
		h.buf.Reset()
		h.mutex.Unlock()
	}()

	// Ensure the buffer is reset before use
	h.buf.Reset()

	if err := h.jsonHandler.Handle(ctx, r); err != nil {
		return nil, nil, fmt.Errorf("error when calling inner handler's Handle: %w", err)
	}

	cloned := bytes.Clone(h.buf.Bytes())
	return cloned, func() {}, nil
}

// mapToSlogAttrs converts a map to a slice of slog.Attr.
// Depth is the current depth of the map traversal, start with 0.
// maxDepth is the maximum depth to flatten the map to.
// Anything deeper than maxDepth will be serialized as JSON string.
func mapToSlogAttrs(input map[string]any, depth int, maxDepth int) []slog.Attr {
	// flatten the map if the depth is greater than maxDepth
	if depth > maxDepth {
		indented, err := marshalIndent(input)
		if err != nil {
			raw, rawErr := json.Marshal(input)
			if rawErr != nil {
				return []slog.Attr{slog.String("data", fmt.Sprintf("error marshalling to JSON: %v", err))}
			}
			return []slog.Attr{slog.String("data", string(raw))}
		}
		return []slog.Attr{slog.String("data", string(indented))}
	}

	var attrs []slog.Attr
	for k, v := range input {
		switch val := v.(type) {
		case map[string]any:
			nestedAttrs := mapToSlogAttrs(val, depth+1, maxDepth)
			// Convert []slog.Attr to []any
			nestedAny := lo.Map(nestedAttrs, func(attr slog.Attr, _ int) any {
				return attr
			})
			attrs = append(attrs, slog.Group(k, nestedAny...))
		default:
			attrs = append(attrs, slog.Any(k, val))
		}
	}
	return attrs
}

// Handle flattens JSON fields into slog attributes before delegating to the
// wrapped handler.
func (h *MyWrapperHandler) Handle(ctx context.Context, r slog.Record) error {
	// Create a new record with the flattened JSON
	newRecord := r
	// Get the JSON bytes from the inner handler
	jsonBytes, release, err := h.GetJsonAttrBytes(ctx, r)
	if err != nil {
		return err
	}
	defer release()

	var attrs map[string]any
	err = json.Unmarshal(jsonBytes, &attrs)
	if err != nil {
		return fmt.Errorf("error when unmarshaling json bytes: %w", err)
	}
	flatAttrs := mapToSlogAttrs(attrs, 0, 5)

	newRecord.AddAttrs(flatAttrs...)

	return h.handler.Handle(ctx, newRecord)
}

// WithAttrs returns a new handler with the provided attributes.
func (h *MyWrapperHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &MyWrapperHandler{
		handler:     h.handler.WithAttrs(attrs),
		jsonHandler: h.jsonHandler.WithAttrs(attrs),
		buf:         h.buf,
		mutex:       h.mutex,
	}
}

// WithGroup returns a new handler with the provided group name.
func (h *MyWrapperHandler) WithGroup(name string) slog.Handler {
	return &MyWrapperHandler{
		handler:     h.handler.WithGroup(name),
		jsonHandler: h.jsonHandler.WithGroup(name),
		buf:         h.buf,
		mutex:       h.mutex,
	}
}

func marshalIndent(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}
