package automapper

import (
	"hwexp/internal/model"
	"testing"
	"time"
)

func device(vendor, class string) model.DiscoveredDevice {
	return model.DiscoveredDevice{
		Platform:    "linux",
		Source:      "linux_hwmon",
		Vendor:      vendor,
		DeviceClass: class,
		FirstSeen:   time.Now(),
		LastSeen:    time.Now(),
	}
}

func raw(name, unit, component, sensor string) model.RawMeasurement {
	return model.RawMeasurement{
		RawName:       name,
		RawUnit:       unit,
		ComponentHint: component,
		SensorHint:    sensor,
		Timestamp:     time.Now(),
	}
}

func TestInferRule_Temperature(t *testing.T) {
	rule := InferRule(
		device("amd", "gpu"),
		raw("temp1_input", "millidegree_celsius", "thermal", "temperature"),
	)
	if rule == nil {
		t.Fatal("expected a rule, got nil")
	}
	if rule.ID != "auto_amd_gpu_temp" {
		t.Errorf("unexpected rule ID: %s", rule.ID)
	}
	if rule.Normalize.MetricFamily != "hw_device_temperature_celsius" {
		t.Errorf("unexpected metric family: %s", rule.Normalize.MetricFamily)
	}
	if rule.Normalize.UnitScale != 0.001 {
		t.Errorf("unexpected unit scale: %f", rule.Normalize.UnitScale)
	}
	if rule.Priority != 1 {
		t.Errorf("expected priority 1 (lowest), got %d", rule.Priority)
	}
}

func TestInferRule_FanRPM(t *testing.T) {
	rule := InferRule(
		device("", "motherboard"),
		raw("fan2_input", "rpm", "cooling", "fan_speed"),
	)
	if rule == nil {
		t.Fatal("expected a rule, got nil")
	}
	if rule.Normalize.MetricFamily != "hw_device_fan_rpm" {
		t.Errorf("unexpected metric family: %s", rule.Normalize.MetricFamily)
	}
	if rule.Normalize.UnitScale != 1.0 {
		t.Errorf("unexpected unit scale: %f", rule.Normalize.UnitScale)
	}
}

func TestInferRule_Power(t *testing.T) {
	rule := InferRule(
		device("", "gpu"),
		raw("power1_input", "microwatt", "power", "power"),
	)
	if rule == nil {
		t.Fatal("expected a rule, got nil")
	}
	if rule.Normalize.MetricFamily != "hw_device_power_watts" {
		t.Errorf("unexpected metric family: %s", rule.Normalize.MetricFamily)
	}
	if rule.Normalize.UnitScale != 0.000001 {
		t.Errorf("unexpected unit scale: %f", rule.Normalize.UnitScale)
	}
}

func TestInferRule_UnknownUnit(t *testing.T) {
	rule := InferRule(
		device("", "sensor"),
		raw("something1_input", "unknown_unit", "", ""),
	)
	if rule != nil {
		t.Errorf("expected nil for unknown unit, got rule %s", rule.ID)
	}
}

func TestInferRule_NonInputName(t *testing.T) {
	rule := InferRule(
		device("amd", "gpu"),
		raw("temp1_max", "millidegree_celsius", "thermal", "temperature"),
	)
	if rule != nil {
		t.Errorf("expected nil for non-input name, got rule %s", rule.ID)
	}
}

func TestInferRule_ComponentHintPreferred(t *testing.T) {
	rule := InferRule(
		device("intel", "cpu"),
		raw("temp1_input", "millidegree_celsius", "custom_component", "custom_sensor"),
	)
	if rule == nil {
		t.Fatal("expected a rule, got nil")
	}
	if rule.Normalize.Component != "custom_component" {
		t.Errorf("expected ComponentHint to be used, got: %s", rule.Normalize.Component)
	}
	if rule.Normalize.Sensor != "custom_sensor" {
		t.Errorf("expected SensorHint to be used, got: %s", rule.Normalize.Sensor)
	}
}
