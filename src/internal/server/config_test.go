package server

import "testing"

func TestConfigFields(t *testing.T) {
	cfg := Config{
		Address:          ":8080",
		BaseURL:          "/api/v1",
		DocsTemplatePath: "/templates/docs.html",
	}

	if cfg.Address != ":8080" {
		t.Errorf("Address = %q, want %q", cfg.Address, ":8080")
	}
	if cfg.BaseURL != "/api/v1" {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, "/api/v1")
	}
	if cfg.DocsTemplatePath != "/templates/docs.html" {
		t.Errorf("DocsTemplatePath = %q, want %q", cfg.DocsTemplatePath, "/templates/docs.html")
	}
}

func TestConfigZeroValue(t *testing.T) {
	var cfg Config

	if cfg.Address != "" {
		t.Errorf("zero value Address = %q, want empty", cfg.Address)
	}
	if cfg.BaseURL != "" {
		t.Errorf("zero value BaseURL = %q, want empty", cfg.BaseURL)
	}
	if cfg.DocsTemplatePath != "" {
		t.Errorf("zero value DocsTemplatePath = %q, want empty", cfg.DocsTemplatePath)
	}
}
