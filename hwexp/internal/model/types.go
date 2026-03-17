package model

import "time"

// DiscoveredDevice represents one physical or logical device known to the exporter.
type DiscoveredDevice struct {
	StableID          string                 `json:"stable_id"`
	Platform          string                 `json:"platform"`
	Source            string                 `json:"source"`
	DeviceClass       string                 `json:"device_class"`
	DeviceSubclass    string                 `json:"device_subclass,omitempty"`
	Vendor            string                 `json:"vendor"`
	Model             string                 `json:"model"`
	Driver            string                 `json:"driver,omitempty"`
	Bus               string                 `json:"bus,omitempty"`
	Location          string                 `json:"location,omitempty"`
	DisplayName       string                 `json:"display_name,omitempty"`
	LogicalDeviceName string                 `json:"logical_device_name,omitempty"`
	Capabilities      []string               `json:"capabilities"`
	RawIdentifiers    map[string]string      `json:"raw_identifiers,omitempty"`
	AdapterMetadata   map[string]interface{} `json:"adapter_metadata,omitempty"`
	FirstSeen         time.Time              `json:"first_seen"`
	LastSeen          time.Time              `json:"last_seen"`
	Present           bool                   `json:"present"`
}

// RawMeasurement represents a sensor reading before normalization.
type RawMeasurement struct {
	MeasurementID  string            `json:"measurement_id"`
	StableDeviceID string            `json:"stable_device_id"`
	Source         string            `json:"source"`
	RawName        string            `json:"raw_name"`
	RawValue       float64           `json:"raw_value"`
	RawUnit        string            `json:"raw_unit"`
	Timestamp      time.Time         `json:"timestamp"`
	Quality        string            `json:"quality"`
	ComponentHint  string            `json:"component_hint,omitempty"`
	SensorHint     string            `json:"sensor_hint,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// NormalizedMeasurement represents a mapped measurement ready for exposure.
type NormalizedMeasurement struct {
	StableDeviceID string            `json:"stable_device_id"`
	LogicalName    string            `json:"logical_name"`
	MetricFamily   string            `json:"metric_family"`
	MetricType     string            `json:"metric_type"`
	Value          float64           `json:"value"`
	Unit           string            `json:"unit"`
	Labels         map[string]string `json:"labels"`
	Quality        string            `json:"quality"`
	MappingRuleID  string            `json:"mapping_rule_id,omitempty"`
	Timestamp      time.Time         `json:"timestamp"`
}
