package httpapi

import (
	"encoding/json"
	"hwexp/internal/config"
	"hwexp/internal/engine"
	"hwexp/internal/model"
	"hwexp/internal/store"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServer_HandleReadyz(t *testing.T) {
	s := store.NewStateStore()
	cfg := &config.Config{}
	srv := NewServer(cfg, s, nil, nil)

	// Test Not Ready
	req := httptest.NewRequest("GET", "/readyz", nil)
	rr := httptest.NewRecorder()
	srv.HandleReadyz(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", rr.Code)
	}

	// Test Ready
	s.UpdateSnapshot(nil, nil, nil, nil)
	rr = httptest.NewRecorder()
	srv.HandleReadyz(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestServer_HandleMetrics(t *testing.T) {
	s := store.NewStateStore()
	cfg := &config.Config{
		Identity: config.IdentityConfig{Host: "test-host"},
	}
	e := engine.NewEngine(s, nil, nil)
	srv := NewServer(cfg, s, e, nil)

	measurements := map[string]model.NormalizedMeasurement{
		"m1": {
			MetricFamily: "hw_device_temperature_celsius",
			MetricType:   "gauge",
			Value:        54.2,
			Labels:       map[string]string{"sensor": "temp1"},
		},
	}
	s.UpdateSnapshot(nil, measurements, nil, nil)

	req := httptest.NewRequest("GET", "/metrics", nil)
	rr := httptest.NewRecorder()
	srv.HandleMetrics(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, `hw_device_temperature_celsius{host="test-host",sensor="temp1"} 54.200000`) {
		t.Errorf("Metrics output missing expected measurement:\n%s", body)
	}
	if !strings.Contains(body, "hwexp_adapter_refresh_duration_seconds") {
		t.Errorf("Metrics output missing self-metrics:\n%s", body)
	}
}

func TestServer_HandleDebugCatalog(t *testing.T) {
	s := store.NewStateStore()
	cfg := &config.Config{
		Identity: config.IdentityConfig{Host: "test-host"},
	}
	srv := NewServer(cfg, s, nil, nil)

	measurements := map[string]model.NormalizedMeasurement{
		"m1": {
			MetricFamily: "hw_device_temperature_celsius",
			Value:        54.2,
		},
	}
	s.UpdateSnapshot(nil, measurements, nil, nil)

	req := httptest.NewRequest("GET", "/debug/catalog", nil)
	rr := httptest.NewRecorder()
	srv.HandleDebugCatalog(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	items := resp["items"].([]interface{})
	if len(items) != 1 {
		t.Errorf("Expected 1 item in catalog, got %d", len(items))
	}
}
