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

func TestEngine_Map_TemplateExpansionMetadataAndRegex(t *testing.T) {
	rules := []model.MappingRule{
		{
			ID:       "gateway_passthrough",
			Priority: 100,
			Match: model.MatchCriteria{
				Source:       "gateway_manifest",
				RawNameRegex: `^(gateway_.+)$`,
			},
			Normalize: model.NormalizeConfig{
				MetricFamily:        "${0}",
				MetricType:          "gauge",
				LogicalNameTemplate: "${logical_device_name}_${stable_device_id}_${zone}",
				Sensor:              "${1}",
				Component:           "${logical_device_name}",
				Labels: map[string]string{
					"copy_raw": "${0}",
					"zone":     "${zone}",
				},
			},
		},
	}

	engine, err := NewEngine(rules)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}

	device := model.DiscoveredDevice{
		StableID:          "gpu-1234",
		Source:            "gateway_manifest",
		LogicalDeviceName: "llm_gateway",
	}

	raw := model.RawMeasurement{
		MeasurementID: "test2",
		RawName:       "gateway_tokens_total",
		RawValue:      42,
		Quality:       "good",
		Metadata: map[string]string{
			"zone": "core0",
		},
		Timestamp: time.Now(),
	}

	norm, decision := engine.Map(device, raw)
	if decision.Decision != "mapped" {
		t.Fatalf("expected decision 'mapped', got %q", decision.Decision)
	}
	if norm == nil {
		t.Fatal("expected normalized measurement, got nil")
	}

	if got, want := norm.MetricFamily, "gateway_tokens_total"; got != want {
		t.Fatalf("expected metric family %q, got %q", want, got)
	}

	if got, want := norm.LogicalName, "llm_gateway_gpu-1234_core0"; got != want {
		t.Fatalf("expected logical name %q, got %q", want, got)
	}

	if got, want := norm.Labels["zone"], "core0"; got != want {
		t.Fatalf("expected label zone=%q, got %q", want, got)
	}

	if got, want := norm.Labels["copy_raw"], "gateway_tokens_total"; got != want {
		t.Fatalf("expected label copy_raw=%q, got %q", want, got)
	}
}
