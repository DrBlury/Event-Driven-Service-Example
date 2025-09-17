package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
)

// MyWrapperHandler is a custom handler that wraps another slog.Handler.
type MyWrapperHandler struct {
	handler     slog.Handler
	jsonHandler slog.Handler
	buf         *bytes.Buffer // Changed to a pointer
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

func (h *MyWrapperHandler) GetJsonAttrBytes(
	ctx context.Context,
	r slog.Record,
) ([]byte, error) {
	h.mutex.Lock()
	defer func() {
		h.buf.Reset()
		h.mutex.Unlock()
	}()

	// Ensure the buffer is reset before use
	h.buf.Reset()

	if err := h.jsonHandler.Handle(ctx, r); err != nil {
		return nil, fmt.Errorf("error when calling inner handler's Handle: %w", err)
	}

	return h.buf.Bytes(), nil
}

// mapToSlogAttrs converts a map to a slice of slog.Attr.
// Depth is the current depth of the map traversal, start with 0.
// maxDepth is the maximum depth to flatten the map to.
// Anything deeper than maxDepth will be serialized as JSON string.
func mapToSlogAttrs(input map[string]any, depth int, maxDepth int) []slog.Attr {
	// flatten the map if the depth is greater than maxDepth
	if depth > maxDepth {
		// Serialize the entire map as JSON string
		jsonBytes, err := json.Marshal(input)
		if err != nil {
			return []slog.Attr{slog.String("data", fmt.Sprintf("error marshalling to JSON: %v", err))}
		}
		// do identation
		identedJSON, err := json.MarshalIndent(input, "", "  ")
		if err != nil {
			return []slog.Attr{slog.String("data", string(jsonBytes))}
		}
		return []slog.Attr{slog.String("data", string(identedJSON))}
	}

	var attrs []slog.Attr
	for k, v := range input {
		switch val := v.(type) {
		case map[string]any:
			nestedAttrs := mapToSlogAttrs(val, depth+1, maxDepth)
			// Convert []slog.Attr to []any
			nestedAny := make([]any, len(nestedAttrs))
			for i, attr := range nestedAttrs {
				nestedAny[i] = attr
			}
			attrs = append(attrs, slog.Group(k, nestedAny...))
		default:
			attrs = append(attrs, slog.Any(k, val))
		}
	}
	return attrs
}

// Handle modifies the record and passes it to the wrapped handler.
func (h *MyWrapperHandler) Handle(ctx context.Context, r slog.Record) error {
	// Create a new record with the flattened JSON
	newRecord := r
	// Get the JSON bytes from the inner handler
	jsonBytes, err := h.GetJsonAttrBytes(ctx, r)
	if err != nil {
		return err
	}

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
