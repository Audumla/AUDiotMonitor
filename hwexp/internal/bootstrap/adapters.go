package bootstrap

import (
	"fmt"
	"time"

	"hwexp/internal/adapters/gateway_manifest"
	"hwexp/internal/adapters/hwmon"
	"hwexp/internal/adapters/linux_gpu"
	"hwexp/internal/adapters/linux_static"
	"hwexp/internal/adapters/linux_storage"
	"hwexp/internal/adapters/linux_system"
	"hwexp/internal/adapters/mock"
	"hwexp/internal/adapters/vendor_exec"
	"hwexp/internal/config"
	"hwexp/internal/engine"
)

// EventLogger is the minimal logger contract required during startup wiring.
type EventLogger interface {
	Info(event, message string, details map[string]interface{})
}

// BuildAdapters builds the runtime adapter set from config and optional fixture.
func BuildAdapters(cfg *config.Config, fixturePath string, l EventLogger) ([]engine.Adapter, error) {
	var adapters []engine.Adapter

	// Fixture mode is explicit and exclusive for deterministic tests.
	if fixturePath != "" {
		if l != nil {
			l.Info("startup", "Using mock adapter (--fixture override)", map[string]interface{}{"fixture": fixturePath})
		}
		a := mock.NewAdapter(fixturePath)
		if err := a.Load(); err != nil {
			return nil, fmt.Errorf("failed to load fixture: %w", err)
		}
		return append(adapters, a), nil
	}

	if cfg.Adapters.LinuxHwmon.Enabled {
		hwmonPath := hwmon.DefaultBasePath
		if p, ok := cfg.Adapters.LinuxHwmon.Settings["hwmon_path"].(string); ok && p != "" {
			hwmonPath = p
		}
		if l != nil {
			l.Info("startup", "Using linux_hwmon adapter", map[string]interface{}{"path": hwmonPath})
		}
		adapters = append(adapters, hwmon.NewAdapter(hwmonPath))
	}

	if cfg.Adapters.LinuxGpuVendor.Enabled {
		if l != nil {
			l.Info("startup", "Using linux_gpu adapter", nil)
		}
		adapters = append(adapters, linux_gpu.NewAdapter(""))
	}

	if cfg.Adapters.LinuxStorage.Enabled {
		if l != nil {
			l.Info("startup", "Using linux_storage adapter", nil)
		}
		adapters = append(adapters, linux_storage.NewAdapter())
	}

	if cfg.Adapters.LinuxSystem.Enabled {
		if l != nil {
			l.Info("startup", "Using linux_system adapter", nil)
		}
		adapters = append(adapters, linux_system.NewAdapter())
	}

	// linux_static is always enabled when not in fixture mode.
	if l != nil {
		l.Info("startup", "Using linux_static adapter", nil)
	}
	adapters = append(adapters, linux_static.NewAdapter(""))

	if cfg.Adapters.LinuxVendorExec.Enabled {
		scriptsDir := "/etc/hwexp/custom.d"
		if d, ok := cfg.Adapters.LinuxVendorExec.Settings["scripts_dir"].(string); ok && d != "" {
			scriptsDir = d
		}
		sourceFormat := "json"
		if f, ok := cfg.Adapters.LinuxVendorExec.Settings["source_format"].(string); ok && f != "" {
			sourceFormat = f
		}
		timeout := cfg.Adapters.LinuxVendorExec.Timeout
		if timeout == 0 {
			timeout = 5 * time.Second
		}

		if l != nil {
			l.Info("startup", "Using vendor_exec adapter", map[string]interface{}{
				"dir":           scriptsDir,
				"timeout":       timeout.String(),
				"source_format": sourceFormat,
			})
		}
		adapters = append(adapters, vendor_exec.NewAdapter(scriptsDir, timeout, sourceFormat))
	}

	if cfg.Adapters.GatewayManifest.Enabled {
		projectDir := "/etc/hwexp/components"
		if d, ok := cfg.Adapters.GatewayManifest.Settings["manifest_dir"].(string); ok && d != "" {
			projectDir = d
		}
		localDir := "/etc/hwexp/local/components"
		if d, ok := cfg.Adapters.GatewayManifest.Settings["local_manifest_dir"].(string); ok && d != "" {
			localDir = d
		}

		if l != nil {
			l.Info("startup", "Using gateway_manifest adapter", map[string]interface{}{
				"project_dir": projectDir,
				"local_dir":   localDir,
			})
		}
		adapters = append(adapters, gateway_manifest.NewAdapter(projectDir, localDir))
	}

	if len(adapters) == 0 {
		return nil, fmt.Errorf("no adapters enabled in config")
	}

	return adapters, nil
}
