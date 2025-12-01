package events

import (
	"testing"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		DemoConsumeQueue:    "messages",
		DemoPublishQueue:    "messages-processed",
		ExampleConsumeQueue: "example-records",
		ExamplePublishQueue: "example-records-processed",
	}

	if cfg.DemoConsumeQueue != "messages" {
		t.Errorf("DemoConsumeQueue = %q, want 'messages'", cfg.DemoConsumeQueue)
	}
	if cfg.DemoPublishQueue != "messages-processed" {
		t.Errorf("DemoPublishQueue = %q, want 'messages-processed'", cfg.DemoPublishQueue)
	}
	if cfg.ExampleConsumeQueue != "example-records" {
		t.Errorf("ExampleConsumeQueue = %q, want 'example-records'", cfg.ExampleConsumeQueue)
	}
	if cfg.ExamplePublishQueue != "example-records-processed" {
		t.Errorf("ExamplePublishQueue = %q, want 'example-records-processed'", cfg.ExamplePublishQueue)
	}
}

func TestConfigEmpty(t *testing.T) {
	t.Parallel()

	cfg := &Config{}

	if cfg.DemoConsumeQueue != "" {
		t.Errorf("DemoConsumeQueue should be empty")
	}
	if cfg.DemoPublishQueue != "" {
		t.Errorf("DemoPublishQueue should be empty")
	}
	if cfg.ExampleConsumeQueue != "" {
		t.Errorf("ExampleConsumeQueue should be empty")
	}
	if cfg.ExamplePublishQueue != "" {
		t.Errorf("ExamplePublishQueue should be empty")
	}
}
