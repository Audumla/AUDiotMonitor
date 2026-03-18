package engine

import (
	"context"
	"hwexp/internal/mapper"
	"hwexp/internal/model"
	"hwexp/internal/store"
	"testing"
)

type mockAdapter struct {
	devices []model.DiscoveredDevice
	raws    []model.RawMeasurement
}

func (m *mockAdapter) Discover(ctx context.Context) ([]model.DiscoveredDevice, error) {
	return m.devices, nil
}

func (m *mockAdapter) Poll(ctx context.Context) ([]model.RawMeasurement, error) {
	return m.raws, nil
}

func TestEngine_executeCycle(t *testing.T) {
	s := store.NewStateStore()
	m, _ := mapper.NewEngine(nil)
	
	adapter := &mockAdapter{
		devices: []model.DiscoveredDevice{
			{StableID: "dev1", DeviceClass: "gpu", Vendor: "amd", Platform: "linux", Source: "test"},
		},
		raws: []model.RawMeasurement{
			{MeasurementID: "m1", StableDeviceID: "dev1", RawName: "temp1_input", RawValue: 50000, RawUnit: "millidegree_celsius"},
		},
	}

	e := NewEngine(s, m, []Adapter{adapter})
	e.EnableAutoMap("") // Enable automap to handle the unmapped sensor

	e.executeCycle(context.Background())
	// Second cycle should now pick up the inferred rule
	e.executeCycle(context.Background())

	if !s.IsReady() {
		t.Error("Store should be ready after one cycle")
	}

	measurements := s.GetAllNormalized()
	if len(measurements) != 1 {
		t.Fatalf("Expected 1 normalized measurement, got %d", len(measurements))
	}

	if measurements[0].Value != 50.0 {
		t.Errorf("Expected value 50.0, got %f", measurements[0].Value)
	}

	metrics := e.GetSelfMetrics()
	if metrics.DiscoveredDevices != 1 {
		t.Errorf("Expected 1 discovered device in metrics, got %d", metrics.DiscoveredDevices)
	}
	if !metrics.LastRefreshSuccess {
		t.Error("Expected LastRefreshSuccess to be true")
	}
}
