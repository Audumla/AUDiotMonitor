package model

type Manifest struct {
	ID          string               `yaml:"id"`
	DisplayName string               `yaml:"display_name"`
	Enabled     bool                 `yaml:"enabled"`
	Health      HealthConfig         `yaml:"health"`
	Discovery   *DiscoveryConfig     `yaml:"discovery,omitempty"`
	Metrics     []MetricConfig       `yaml:"metrics"`
	Connection  ConnectionConfig     `yaml:"connection"`
	Correlation *HardwareCorrelation `yaml:"hardware_correlation,omitempty"`
}

type DiscoveryConfig struct {
	Type             string `yaml:"type"` // e.g., "llama-swap"
	Endpoint         string `yaml:"endpoint"`
	ActivityField    string `yaml:"activity_field"`    // e.g., "requests_processing"
	BackendPortField string `yaml:"backend_port_field"` // e.g., "port"
}

type HealthConfig struct {
	Endpoint     string `yaml:"endpoint"`
	ExpectStatus int    `yaml:"expect_status"`
	TimeoutS     int    `yaml:"timeout_s"`
}

type MetricConfig struct {
	ID             string `yaml:"id"`
	SourceType     string `yaml:"source_type"` // "http" (default), "exec", "file"
	Endpoint       string `yaml:"endpoint"`    // URL path, command, or file path
	Extract        string `yaml:"extract"`
	PrometheusName string `yaml:"prometheus_name"`
	Unit           string `yaml:"unit"`
	PollIntervalS  int    `yaml:"poll_interval_s"`
	SourceFormat   string `yaml:"source_format"` // "json" or "prometheus"
}

type ConnectionConfig struct {
	Host string     `yaml:"host"`
	Port int        `yaml:"port"`
	Auth AuthConfig `yaml:"auth"`
}

type AuthConfig struct {
	Type     string `yaml:"type"` // "none" or "bearer"
	TokenEnv string `yaml:"token_env"`
}

type HardwareCorrelation struct {
	DeviceClass string `yaml:"device_class"`
	PCISlot     string `yaml:"pci_slot"`
}
