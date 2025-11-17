package logging

import "drblury/event-driven-service/pkg/jsonutil"

// marshalIndent produces deterministically indented JSON output using json/v2 semantics.
func marshalIndent(v any) ([]byte, error) {
	return jsonutil.MarshalIndent(v, "", "  ")
}
