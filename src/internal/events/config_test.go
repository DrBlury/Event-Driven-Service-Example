package events

import "testing"

func TestEventsConfigFields(t *testing.T) {
	cfg := Config{
		DemoConsumeQueue:    "demo-consume",
		DemoPublishQueue:    "demo-publish",
		ExampleConsumeQueue: "example-consume",
		ExamplePublishQueue: "example-publish",
	}

	if cfg.DemoConsumeQueue != "demo-consume" {
		t.Errorf("DemoConsumeQueue = %q, want %q", cfg.DemoConsumeQueue, "demo-consume")
	}
	if cfg.DemoPublishQueue != "demo-publish" {
		t.Errorf("DemoPublishQueue = %q, want %q", cfg.DemoPublishQueue, "demo-publish")
	}
	if cfg.ExampleConsumeQueue != "example-consume" {
		t.Errorf("ExampleConsumeQueue = %q, want %q", cfg.ExampleConsumeQueue, "example-consume")
	}
	if cfg.ExamplePublishQueue != "example-publish" {
		t.Errorf("ExamplePublishQueue = %q, want %q", cfg.ExamplePublishQueue, "example-publish")
	}
}

func TestEventsConfigZeroValue(t *testing.T) {
	var cfg Config

	if cfg.DemoConsumeQueue != "" {
		t.Errorf("zero value DemoConsumeQueue = %q, want empty string", cfg.DemoConsumeQueue)
	}
	if cfg.DemoPublishQueue != "" {
		t.Errorf("zero value DemoPublishQueue = %q, want empty string", cfg.DemoPublishQueue)
	}
	if cfg.ExampleConsumeQueue != "" {
		t.Errorf("zero value ExampleConsumeQueue = %q, want empty string", cfg.ExampleConsumeQueue)
	}
	if cfg.ExamplePublishQueue != "" {
		t.Errorf("zero value ExamplePublishQueue = %q, want empty string", cfg.ExamplePublishQueue)
	}
}
