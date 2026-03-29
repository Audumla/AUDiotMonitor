package bootstrap

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"hwexp/internal/config"
)

type testLogger struct{}

func (testLogger) Info(event, message string, details map[string]interface{}) {}

func TestBuildAdapters_FixtureModeExclusive(t *testing.T) {
	tmpDir := t.TempDir()
	fixturePath := filepath.Join(tmpDir, "fixture.json")
	if err := os.WriteFile(fixturePath, []byte(`{"devices":[],"measurements":[]}`), 0644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	cfg := &config.Config{}
	adapters, err := BuildAdapters(cfg, fixturePath, testLogger{})
	if err != nil {
		t.Fatalf("BuildAdapters returned error: %v", err)
	}
	if len(adapters) != 1 {
		t.Fatalf("expected exactly one adapter in fixture mode, got %d", len(adapters))
	}
	if !strings.Contains(reflect.TypeOf(adapters[0]).String(), "mock.Adapter") {
		t.Fatalf("expected mock adapter in fixture mode, got %T", adapters[0])
	}
}

func TestBuildAdapters_ConfigModeIncludesStaticAndEnabled(t *testing.T) {
	cfg := &config.Config{
		Adapters: config.AdaptersConfig{
			LinuxStorage: config.AdapterConfig{
				Enabled: true,
			},
			LinuxVendorExec: config.AdapterConfig{
				Enabled: true,
				Timeout: 3 * time.Second,
				Settings: map[string]interface{}{
					"scripts_dir":   "/opt/scripts",
					"source_format": "prometheus",
				},
			},
		},
	}

	adapters, err := BuildAdapters(cfg, "", testLogger{})
	if err != nil {
		t.Fatalf("BuildAdapters returned error: %v", err)
	}

	typeNames := make([]string, 0, len(adapters))
	for _, a := range adapters {
		typeNames = append(typeNames, reflect.TypeOf(a).String())
	}
	joined := strings.Join(typeNames, ",")
	if !strings.Contains(joined, "linux_static.Adapter") {
		t.Fatalf("expected linux_static adapter, got %s", joined)
	}
	if !strings.Contains(joined, "linux_storage.Adapter") {
		t.Fatalf("expected linux_storage adapter, got %s", joined)
	}
	if !strings.Contains(joined, "vendor_exec.Adapter") {
		t.Fatalf("expected vendor_exec adapter, got %s", joined)
	}
}
