package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestMarshal(t *testing.T) {
	t.Parallel()

	data := map[string]string{"key": "value"}
	result, err := Marshal(data)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if !bytes.Contains(result, []byte("key")) {
		t.Error("Marshal result should contain 'key'")
	}
}

func TestMarshalIndent(t *testing.T) {
	t.Parallel()

	data := map[string]string{"key": "value"}
	result, err := MarshalIndent(data, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent failed: %v", err)
	}
	if !bytes.Contains(result, []byte("key")) {
		t.Error("MarshalIndent result should contain 'key'")
	}
	if !bytes.Contains(result, []byte("  ")) {
		t.Error("MarshalIndent result should contain indentation")
	}
}

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	data := []byte(`{"key": "value"}`)
	var result map[string]string
	err := Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("Unmarshal result = %v, want map with key=value", result)
	}
}

func TestUnmarshalInvalid(t *testing.T) {
	t.Parallel()

	data := []byte(`{invalid}`)
	var result map[string]string
	err := Unmarshal(data, &result)
	if err == nil {
		t.Error("Unmarshal should fail with invalid JSON")
	}
}

func TestEncode(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	data := map[string]string{"key": "value"}
	err := Encode(buf, data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if !strings.Contains(buf.String(), "key") {
		t.Error("Encoded result should contain 'key'")
	}
}

func TestDecode(t *testing.T) {
	t.Parallel()

	data := strings.NewReader(`{"key": "value"}`)
	var result map[string]string
	err := Decode(data, &result)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("Decode result = %v, want map with key=value", result)
	}
}

func TestDecodeInvalid(t *testing.T) {
	t.Parallel()

	data := strings.NewReader(`{invalid}`)
	var result map[string]string
	err := Decode(data, &result)
	if err == nil {
		t.Error("Decode should fail with invalid JSON")
	}
}

func TestMarshalNil(t *testing.T) {
	t.Parallel()

	result, err := Marshal(nil)
	if err != nil {
		t.Fatalf("Marshal nil failed: %v", err)
	}
	if string(result) != "null" {
		t.Errorf("Marshal nil = %q, want 'null'", string(result))
	}
}

func TestMarshalComplexType(t *testing.T) {
	t.Parallel()

	type nested struct {
		Name  string   `json:"name"`
		Items []string `json:"items"`
	}
	data := nested{
		Name:  "test",
		Items: []string{"a", "b", "c"},
	}
	result, err := Marshal(data)
	if err != nil {
		t.Fatalf("Marshal complex type failed: %v", err)
	}
	if !bytes.Contains(result, []byte("test")) {
		t.Error("Marshal result should contain 'test'")
	}
	if !bytes.Contains(result, []byte("items")) {
		t.Error("Marshal result should contain 'items'")
	}
}
