package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Identity IdentityConfig `yaml:"identity"`
	Adapters AdaptersConfig `yaml:"adapters"`
	Mapping  MappingConfig  `yaml:"mapping"`
	Filters  FiltersConfig  `yaml:"filters,omitempty"`
	Security SecurityConfig `yaml:"security,omitempty"`
	Debug    DebugConfig    `yaml:"debug,omitempty"`
}

type ServerConfig struct {
	ListenAddress             string        `yaml:"listen_address"`
	RefreshInterval           time.Duration `yaml:"refresh_interval"`
	DiscoveryInterval         time.Duration `yaml:"discovery_interval"`
	GracePeriod               time.Duration `yaml:"grace_period"`
	RequestTimeout            time.Duration `yaml:"request_timeout"`
	MaxConcurrentAdapterPolls int           `yaml:"max_concurrent_adapter_polls"`
}

type IdentityConfig struct {
	Host         string            `yaml:"host"`
	Platform     string            `yaml:"platform"`
	Site         string            `yaml:"site,omitempty"`
	Role         string            `yaml:"role,omitempty"`
	Environment  string            `yaml:"environment,omitempty"`
	StaticLabels map[string]string `yaml:"static_labels,omitempty"`
}

type AdaptersConfig struct {
	LinuxHwmon            AdapterConfig `yaml:"linux_hwmon,omitempty"`
	LinuxGpuVendor        AdapterConfig `yaml:"linux_gpu_vendor,omitempty"`
	LinuxVendorExec       AdapterConfig `yaml:"linux_vendor_exec,omitempty"`
	LinuxNodeBridge       AdapterConfig `yaml:"linux_node_bridge,omitempty"`
	WindowsExporterBridge AdapterConfig `yaml:"windows_exporter_bridge,omitempty"`
	DarwinNodeBridge      AdapterConfig `yaml:"darwin_node_bridge,omitempty"`
}

type AdapterConfig struct {
	Enabled  bool                   `yaml:"enabled"`
	Priority int                    `yaml:"priority,omitempty"`
	Timeout  time.Duration          `yaml:"timeout,omitempty"`
	Settings map[string]interface{} `yaml:"settings,omitempty"`
}

type MappingConfig struct {
	RulesFile          string `yaml:"rules_file"`
	AliasesFile        string `yaml:"aliases_file,omitempty"`
	StrictMode         bool   `yaml:"strict_mode"`
	DefaultDropUnmapped bool   `yaml:"default_drop_unmapped"`
}

type FiltersConfig struct {
	IncludeDeviceClasses []string `yaml:"include_device_classes,omitempty"`
	ExcludeDeviceClasses []string `yaml:"exclude_device_classes,omitempty"`
	IncludeSources       []string `yaml:"include_sources,omitempty"`
	ExcludeRawNameRegex  []string `yaml:"exclude_raw_name_regex,omitempty"`
	SuppressQualities    []string `yaml:"suppress_qualities,omitempty"`
}

type SecurityConfig struct {
	AuthMode                   string `yaml:"auth_mode"`
	BindScope                  string `yaml:"bind_scope"`
	TLSEnabled                 bool   `yaml:"tls_enabled"`
	TLSCertFile                string `yaml:"tls_cert_file,omitempty"`
	TLSKeyFile                 string `yaml:"tls_key_file,omitempty"`
	APITokensFile              string `yaml:"api_tokens_file,omitempty"`
	DebugEndpointsEnabled      bool   `yaml:"debug_endpoints_enabled"`
	DebugEndpointsAuthRequired bool   `yaml:"debug_endpoints_auth_required"`
}

type DebugConfig struct {
	EnableRawEndpoint        bool   `yaml:"enable_raw_endpoint"`
	LogLevel                 string `yaml:"log_level"`
	RetainLastMappingCycles  int    `yaml:"retain_last_mapping_cycles"`
	RetainLastRawCycles      int    `yaml:"retain_last_raw_cycles"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	// Set defaults
	cfg.Server.RefreshInterval = 5 * time.Second
	cfg.Server.DiscoveryInterval = 60 * time.Second
	cfg.Server.GracePeriod = 30 * time.Second
	cfg.Server.RequestTimeout = 5 * time.Second
	cfg.Server.MaxConcurrentAdapterPolls = 4
	cfg.Identity.Platform = "auto"
	cfg.Security.AuthMode = "none"
	cfg.Security.BindScope = "lan"
	cfg.Security.DebugEndpointsEnabled = true
	cfg.Debug.LogLevel = "info"
	cfg.Debug.RetainLastMappingCycles = 5
	cfg.Debug.RetainLastRawCycles = 1

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Basic validation
	if cfg.Server.ListenAddress == "" {
		return nil, fmt.Errorf("server.listen_address is required")
	}
	if cfg.Identity.Host == "" {
		return nil, fmt.Errorf("identity.host is required")
	}
	if cfg.Mapping.RulesFile == "" {
		return nil, fmt.Errorf("mapping.rules_file is required")
	}

	return cfg, nil
}
