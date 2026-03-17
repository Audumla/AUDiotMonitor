package mapper

import (
	"hwexp/internal/model"
	"testing"
	"time"
)

func TestEngine_Map(t *testing.T) {
	rules := []model.MappingRule{
		{
			ID:       "amd_gpu_temp",
			Priority: 100,
			Match: model.MatchCriteria{
				Vendor:       "amd",
				RawNameRegex: `temp([0-9]+)_input`,
			},
			Normalize: model.NormalizeConfig{
				MetricFamily:        "hw_device_temperature_celsius",
				MetricType:          "gauge",
				LogicalNameTemplate: "${logical_device_name}_temp_${1}",
				UnitScale:           0.001,
				Component:           "thermal",
			},
		},
	}

	engine, err := NewEngine(rules)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	device := model.DiscoveredDevice{
		StableID:          "pci-0000",
		Vendor:            "amd",
		LogicalDeviceName: "gpu0",
	}

	raw := model.RawMeasurement{
		MeasurementID: "test1",
		RawName:       "temp1_input",
		RawValue:      54234,
		Quality:       "good",
		Timestamp:     time.Now(),
	}

	norm, decision := engine.Map(device, raw)

	if decision.Decision != "mapped" {
		t.Errorf("Expected decision 'mapped', got '%s'", decision.Decision)
	}

	if norm == nil {
		t.Fatal("Expected normalized measurement, got nil")
	}

	if norm.Value != 54.234 {
		t.Errorf("Expected scaled value 54.234, got %f", norm.Value)
	}

	if norm.LogicalName != "gpu0_temp_1" {
		t.Errorf("Expected logical name gpu0_temp_1, got %s", norm.LogicalName)
	}
	
	if norm.Labels["component"] != "thermal" {
		t.Errorf("Expected label component=thermal, got %s", norm.Labels["component"])
	}
}
