package main

import (
	"os"
	"testing"
)

func TestGetApiBaseUrlReturnsCompany(t *testing.T) {
	tempDir := t.TempDir()

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	cfg := Config{}
	cfg.User.Company = "acme"

	if err := WriteConfig(cfg); err != nil {
		t.Fatalf("WriteConfig error: %v", err)
	}

	baseURL, err := GetApiBaseUrl()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if baseURL != "https://acme.aerion.app" {
		t.Fatalf("expected base url %q, got %q", cfg.User.Company, baseURL)
	}
}

func TestGetApiBaseUrlMissingCompany(t *testing.T) {
	tempDir := t.TempDir()

	oldHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", oldHome)
	})

	if err := WriteConfig(Config{}); err != nil {
		t.Fatalf("WriteConfig error: %v", err)
	}

	if _, err := GetApiBaseUrl(); err == nil {
		t.Fatal("expected error when company is not set")
	}
}
