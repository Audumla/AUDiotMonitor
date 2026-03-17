package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	content := `
server:
  listen_address: "127.0.0.1:9200"
identity:
  host: "test-host"
mapping:
  rules_file: "mappings.yaml"
`
	tmpfile, err := os.CreateTemp("", "hwexp-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Server.ListenAddress != "127.0.0.1:9200" {
		t.Errorf("Expected listen_address 127.0.0.1:9200, got %s", cfg.Server.ListenAddress)
	}
	if cfg.Identity.Host != "test-host" {
		t.Errorf("Expected host test-host, got %s", cfg.Identity.Host)
	}
	if cfg.Server.RefreshInterval != 5*time.Second {
		t.Errorf("Expected default refresh_interval 5s, got %v", cfg.Server.RefreshInterval)
	}
}
