// Package automapper infers MappingRules for unmapped hwmon measurements
// using the Linux kernel hwmon sysfs ABI naming conventions.
// Reference: https://www.kernel.org/doc/Documentation/hwmon/sysfs-interface
// and Prometheus node_exporter hwmon collector.
package automapper

import (
	"fmt"
	"regexp"

	"hwexp/internal/model"
)

// sensorMapping describes how to normalise a sensor type.
type sensorMapping struct {
	metricFamily string
	scale        float64
	component    string
	sensor       string
}

// unitToMapping maps the RawUnit values produced by the hwmon adapter
// to their normalised metric representation.
var unitToMapping = map[string]sensorMapping{
	"millidegree_celsius": {
		metricFamily: "hw_device_temperature_celsius",
		scale:        0.001,
		component:    "thermal",
		sensor:       "temperature",
	},
	"rpm": {
		metricFamily: "hw_device_fan_rpm",
		scale:        1.0,
		component:    "cooling",
		sensor:       "fan_speed",
	},
	"microwatt": {
		metricFamily: "hw_device_power_watts",
		scale:        0.000001,
		component:    "power",
		sensor:       "power",
	},
	"millivolt": {
		metricFamily: "hw_device_voltage_volts",
		scale:        0.001,
		component:    "power",
		sensor:       "voltage",
	},
	"milliampere": {
		metricFamily: "hw_device_current_amps",
		scale:        0.001,
		component:    "power",
		sensor:       "current",
	},
	"hertz": {
		metricFamily: "hw_device_frequency_hz",
		scale:        1.0,
		component:    "compute",
		sensor:       "frequency",
	},
	"millipercent": {
		metricFamily: "hw_device_humidity_percent",
		scale:        0.001,
		component:    "environment",
		sensor:       "humidity",
	},
	"microjoule": {
		metricFamily: "hw_device_energy_joules",
		scale:        0.000001,
		component:    "power",
		sensor:       "energy",
	},
}

// inputRE extracts the sensor prefix (e.g. "temp") from a raw hwmon filename
// like "temp1_input".
var inputRE = regexp.MustCompile(`^([a-z]+)\d+_input$`)

// InferRule creates a best-guess MappingRule for an unmapped measurement.
// The rule is scoped to the device's (vendor, device_class) and matches all
// sensors of the same type on that class of device.
// Returns nil if the unit is not recognised.
func InferRule(device model.DiscoveredDevice, raw model.RawMeasurement) *model.MappingRule {
	sm, ok := unitToMapping[raw.RawUnit]
	if !ok {
		return nil
	}

	// Extract sensor prefix from raw name ("temp1_input" → "temp")
	m := inputRE.FindStringSubmatch(raw.RawName)
	if m == nil {
		return nil
	}
	prefix := m[1]

	// Rule ID is deterministic: same (vendor, class, prefix) → same rule ID,
	// so duplicates are naturally deduplicated when AddRules checks by ID.
	ruleID := fmt.Sprintf("auto_%s_%s_%s", device.Vendor, device.DeviceClass, prefix)

	// Prefer the hint fields set by the adapter, fall back to catalog defaults.
	component := raw.ComponentHint
	if component == "" {
		component = sm.component
	}
	sensor := raw.SensorHint
	if sensor == "" {
		sensor = sm.sensor
	}

	return &model.MappingRule{
		ID:       ruleID,
		Priority: 1, // lowest priority — manual rules always win
		Match: model.MatchCriteria{
			Platform:    device.Platform,
			Source:      device.Source,
			Vendor:      device.Vendor,
			DeviceClass: device.DeviceClass,
			RawNameRegex: fmt.Sprintf(`^%s([0-9]+)_input$`, regexp.QuoteMeta(prefix)),
		},
		Normalize: model.NormalizeConfig{
			MetricFamily:        sm.metricFamily,
			MetricType:          "gauge",
			LogicalNameTemplate: fmt.Sprintf("${logical_device_name}_%s_${1}", prefix),
			UnitScale:           sm.scale,
			DeviceClass:         device.DeviceClass,
			Component:           component,
			Sensor:              sensor,
		},
	}
}
