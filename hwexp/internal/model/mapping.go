package model

// MappingRule defines how a RawMeasurement is transformed into a NormalizedMeasurement.
type MappingRule struct {
	ID        string          `json:"id" yaml:"id"`
	Priority  int             `json:"priority" yaml:"priority"`
	Match     MatchCriteria   `json:"match" yaml:"match"`
	Normalize NormalizeConfig `json:"normalize" yaml:"normalize"`
}

type MatchCriteria struct {
	Platform       string `json:"platform,omitempty" yaml:"platform,omitempty"`
	Source         string `json:"source,omitempty" yaml:"source,omitempty"`
	DeviceClass    string `json:"device_class,omitempty" yaml:"device_class,omitempty"`
	DeviceSubclass string `json:"device_subclass,omitempty" yaml:"device_subclass,omitempty"`
	Vendor         string `json:"vendor,omitempty" yaml:"vendor,omitempty"`
	ModelRegex     string `json:"model_regex,omitempty" yaml:"model_regex,omitempty"`
	StableIDRegex  string `json:"stable_id_regex,omitempty" yaml:"stable_id_regex,omitempty"`
	RawNameRegex   string `json:"raw_name_regex,omitempty" yaml:"raw_name_regex,omitempty"`
	ComponentHint  string `json:"component_hint,omitempty" yaml:"component_hint,omitempty"`
	SensorHint     string `json:"sensor_hint,omitempty" yaml:"sensor_hint,omitempty"`
}

type NormalizeConfig struct {
	MetricFamily        string            `json:"metric_family" yaml:"metric_family"`
	MetricType          string            `json:"metric_type" yaml:"metric_type"`
	DeviceClass         string            `json:"device_class,omitempty" yaml:"device_class,omitempty"`
	DeviceSubclass      string            `json:"device_subclass,omitempty" yaml:"device_subclass,omitempty"`
	Component           string            `json:"component,omitempty" yaml:"component,omitempty"`
	Sensor              string            `json:"sensor,omitempty" yaml:"sensor,omitempty"`
	LogicalNameTemplate string            `json:"logical_name_template" yaml:"logical_name_template"`
	UnitScale           float64           `json:"unit_scale,omitempty" yaml:"unit_scale,omitempty"`
	UnitOffset          float64           `json:"unit_offset,omitempty" yaml:"unit_offset,omitempty"`
	Labels              map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`
	Drop                bool              `json:"drop,omitempty" yaml:"drop,omitempty"`
}

// MappingDecision records how a raw measurement became a normalized measurement or why it was dropped.
type MappingDecision struct {
	MeasurementID  string            `json:"measurement_id"`
	Decision       string            `json:"decision"`
	MappingRuleID  string            `json:"mapping_rule_id,omitempty"`
	Precedence     int               `json:"precedence,omitempty"`
	RawName        string            `json:"raw_name,omitempty"`
	RawUnit        string            `json:"raw_unit,omitempty"`
	ConvertedValue float64           `json:"converted_value,omitempty"`
	MetricFamily   string            `json:"metric_family,omitempty"`
	LogicalName    string            `json:"logical_name,omitempty"`
	Labels         map[string]string `json:"labels,omitempty"`
	Notes          []string          `json:"notes,omitempty"`
}
