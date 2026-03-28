package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"time"

	"hwexp/internal/adapters/gateway_manifest"
	"hwexp/internal/adapters/hwmon"
	"hwexp/internal/adapters/linux_gpu"
	"hwexp/internal/adapters/linux_static"
	"hwexp/internal/adapters/linux_storage"
	"hwexp/internal/adapters/linux_system"
	"hwexp/internal/adapters/mock"
	"hwexp/internal/adapters/vendor_exec"
	"hwexp/internal/capabilities"
	"hwexp/internal/config"
	"hwexp/internal/engine"
	"hwexp/internal/httpapi"
	"hwexp/internal/logger"
	"hwexp/internal/mapper"
	"hwexp/internal/store"
	"hwexp/internal/version"
)

// bannerInner is the number of characters between the two ║ box chars on each line.
const bannerInner = 56

// bannerRow returns a single banner line with the label left-aligned in a fixed
// column and the value truncated (with "…") or padded to fill the remaining space
// exactly — guaranteeing every line is the same total width.
func bannerRow(label, value string) string {
	const labelW = 11                 // width of the label column (e.g. "Host       ")
	const sepW = 2                    // ": "
	const prefixW = 2 + labelW + sepW // "  " + label + ": "
	valueW := bannerInner - prefixW
	if len(value) > valueW {
		value = value[:valueW-1] + "…"
	}
	return fmt.Sprintf(" ║  %-*s: %-*s║", labelW, label, valueW, value)
}

func printBanner(cfg *config.Config, configPath string) {
	// linux_static is always active (not config-gated)
	adapterNames := []string{"linux_static"}
	if cfg.Adapters.LinuxHwmon.Enabled {
		adapterNames = append(adapterNames, "linux_hwmon")
	}
	if cfg.Adapters.LinuxGpuVendor.Enabled {
		adapterNames = append(adapterNames, "linux_gpu_vendor")
	}
	if cfg.Adapters.LinuxVendorExec.Enabled {
		adapterNames = append(adapterNames, "linux_vendor_exec")
	}
	if cfg.Adapters.LinuxStorage.Enabled {
		adapterNames = append(adapterNames, "linux_storage")
	}
	adapterStr := strings.Join(adapterNames, ", ")

	autoMap := "disabled"
	if cfg.Mapping.AutoMap.Enabled {
		autoMap = "enabled"
	}

	border := " ╔" + strings.Repeat("═", bannerInner) + "╗"
	divider := " ╠" + strings.Repeat("═", bannerInner) + "╣"
	footer := " ╚" + strings.Repeat("═", bannerInner) + "╝"

	title := "AUDiot Hardware Exporter  v" + version.Version
	if len(title) > bannerInner {
		title = title[:bannerInner]
	}
	pad := bannerInner - len(title)
	header := " ║" + strings.Repeat(" ", pad/2) + title + strings.Repeat(" ", pad-pad/2) + "║"

	lines := []string{
		"",
		border,
		header,
		divider,
		bannerRow("Host", cfg.Identity.Host),
		bannerRow("Listen", cfg.Server.ListenAddress),
		bannerRow("Config", configPath),
		bannerRow("Refresh", cfg.Server.RefreshInterval.String()),
		bannerRow("Discovery", cfg.Server.DiscoveryInterval.String()),
		bannerRow("Auto-map", autoMap),
		bannerRow("Adapters", adapterStr),
		bannerRow("Go version", version.GoVersion()),
		bannerRow("Build time", version.BuildTime),
		footer,
		"",
	}
	fmt.Fprintln(os.Stderr, strings.Join(lines, "\n"))
}

func main() {
	configPath := flag.String("config", "configs/hwexp.yaml", "Path to config YAML")
	fixturePath := flag.String("fixture", "", "Path to mock fixture JSON (forces mock adapter, overrides config)")
	flag.Parse()

	// 1. Load Config
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. Initialize Logger
	l := logger.New(cfg.Identity.Host, "hwexp")

	// Print human-readable startup banner to stderr before structured logs
	printBanner(cfg, *configPath)

	l.Info("startup", "Starting AUDiot Exporter...", nil)

	// 3. Load Mapping Rules (manual + previously auto-generated)
	rules, err := mapper.LoadRules(cfg.Mapping.RulesFile)
	if err != nil {
		l.Fatal("startup", "Failed to load rules", "CFG_LOAD_FAILED", map[string]interface{}{"error": err.Error()})
	}
	l.Info("startup", "Loaded mapping rules", map[string]interface{}{"count": len(rules), "file": cfg.Mapping.RulesFile})

	if cfg.Mapping.AutoMap.Enabled && cfg.Mapping.AutoMap.GeneratedFile != "" {
		genRules, err := mapper.LoadRules(cfg.Mapping.AutoMap.GeneratedFile)
		if err != nil {
			if !os.IsNotExist(err) {
				// A parse error in the auto-generated file is logged but not fatal;
				// the engine will overwrite the file with fresh rules on next cycle.
				l.Info("startup", "Could not load auto-generated rules (will regenerate)", map[string]interface{}{
					"file":  cfg.Mapping.AutoMap.GeneratedFile,
					"error": err.Error(),
				})
			}
			// Missing file on first run is expected — do nothing.
		} else {
			rules = append(rules, genRules...)
			l.Info("startup", "Loaded auto-generated rules", map[string]interface{}{
				"count": len(genRules),
				"file":  cfg.Mapping.AutoMap.GeneratedFile,
			})
		}
	}

	// 4. Initialize Auth Store
	authStore, err := httpapi.LoadAuthStore(cfg.Security.APITokensFile)
	if err != nil {
		l.Fatal("startup", "Failed to load auth store", "CFG_LOAD_FAILED", map[string]interface{}{"error": err.Error()})
	}

	// 5. Select adapters
	// --fixture flag forces mock mode (for testing/CI).
	// Otherwise use adapters enabled in config.
	var adapters []engine.Adapter

	if *fixturePath != "" {
		l.Info("startup", "Using mock adapter (--fixture override)", map[string]interface{}{"fixture": *fixturePath})
		a := mock.NewAdapter(*fixturePath)
		if err := a.Load(); err != nil {
			l.Fatal("startup", "Failed to load fixture", "ADAPTER_INIT_FAILED", map[string]interface{}{"error": err.Error()})
		}
		adapters = append(adapters, a)
	} else if cfg.Adapters.LinuxHwmon.Enabled {
		hwmonPath := hwmon.DefaultBasePath
		if p, ok := cfg.Adapters.LinuxHwmon.Settings["hwmon_path"].(string); ok && p != "" {
			hwmonPath = p
		}
		l.Info("startup", "Using linux_hwmon adapter", map[string]interface{}{"path": hwmonPath})
		adapters = append(adapters, hwmon.NewAdapter(hwmonPath))
	}

	if cfg.Adapters.LinuxGpuVendor.Enabled && *fixturePath == "" {
		l.Info("startup", "Using linux_gpu adapter", nil)
		adapters = append(adapters, linux_gpu.NewAdapter(""))
	}

	if cfg.Adapters.LinuxStorage.Enabled && *fixturePath == "" {
		l.Info("startup", "Using linux_storage adapter", nil)
		adapters = append(adapters, linux_storage.NewAdapter())
	}

	if cfg.Adapters.LinuxSystem.Enabled && *fixturePath == "" {
		l.Info("startup", "Using linux_system adapter", nil)
		adapters = append(adapters, linux_system.NewAdapter())
	}

	if *fixturePath == "" {
		l.Info("startup", "Using linux_static adapter", nil)
		adapters = append(adapters, linux_static.NewAdapter(""))
	}

	if cfg.Adapters.LinuxVendorExec.Enabled && *fixturePath == "" {
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

		l.Info("startup", "Using vendor_exec adapter", map[string]interface{}{
			"dir":           scriptsDir,
			"timeout":       timeout.String(),
			"source_format": sourceFormat,
		})
		adapters = append(adapters, vendor_exec.NewAdapter(scriptsDir, timeout, sourceFormat))
	}

	if cfg.Adapters.GatewayManifest.Enabled && *fixturePath == "" {
		projectDir := "/etc/hwexp/components"
		if d, ok := cfg.Adapters.GatewayManifest.Settings["manifest_dir"].(string); ok && d != "" {
			projectDir = d
		}
		localDir := "/etc/hwexp/local/components"
		if d, ok := cfg.Adapters.GatewayManifest.Settings["local_manifest_dir"].(string); ok && d != "" {
			localDir = d
		}

		l.Info("startup", "Using gateway_manifest adapter", map[string]interface{}{
			"project_dir": projectDir,
			"local_dir":   localDir,
		})
		adapters = append(adapters, gateway_manifest.NewAdapter(projectDir, localDir))
	}

	if len(adapters) == 0 {
		l.Fatal("startup", "No adapters enabled in config and no --fixture provided", "CFG_LOAD_FAILED", nil)
	}

	// 6. Initialize State Store
	stateStore := store.NewStateStore()
	capabilityProviders := make([]capabilities.Provider, 0, len(adapters))
	for _, a := range adapters {
		if provider, ok := a.(capabilities.Provider); ok {
			capabilityProviders = append(capabilityProviders, provider)
		}
	}
	stateStore.SetCapabilities(capabilities.CheckRequirements(capabilityProviders, nil, l))

	// 7. Initialize Mapping Engine
	mapperEngine, err := mapper.NewEngine(rules)
	if err != nil {
		l.Fatal("startup", "Failed to initialize mapper", "INTERNAL_ERROR", map[string]interface{}{"error": err.Error()})
	}

	// 8. Start Core Engine
	coreEngine := engine.NewEngine(stateStore, mapperEngine, adapters)
	if cfg.Mapping.AutoMap.Enabled {
		coreEngine.EnableAutoMap(cfg.Mapping.AutoMap.GeneratedFile)
		l.Info("startup", "Auto-mapping enabled", map[string]interface{}{"generated_file": cfg.Mapping.AutoMap.GeneratedFile})
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle SIGTERM/SIGINT for graceful shutdown
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
		defer signal.Stop(sigs)
		select {
		case <-ctx.Done():
		case sig := <-sigs:
			l.Info("shutdown", "Received signal, initiating graceful shutdown", map[string]interface{}{"signal": sig.String()})
			cancel()
		}
	}()

	l.Info("startup", "Starting background telemetry loop", map[string]interface{}{"interval": cfg.Server.RefreshInterval})
	coreEngine.Start(ctx)

	// 8b. SIGHUP reloads mapping rules without restart
	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGHUP)
		defer signal.Stop(sigs)
		for {
			select {
			case <-ctx.Done():
				return
			case <-sigs:
				l.Info("reload", "Reloading mapping rules (SIGHUP)", map[string]interface{}{"file": cfg.Mapping.RulesFile})
				if err := mapperEngine.ReloadRules(cfg.Mapping.RulesFile); err != nil {
					l.Info("reload", "Failed to reload mapping rules", map[string]interface{}{"error": err.Error()})
				} else {
					l.Info("reload", "Mapping rules reloaded successfully", nil)
				}
			}
		}
	}()

	// 9. Start HTTP Server — blocks until ctx cancelled or fatal error
	server := httpapi.NewServer(cfg, stateStore, coreEngine, authStore)
	l.Info("startup", "HTTP server listening", map[string]interface{}{"address": cfg.Server.ListenAddress})
	if err := server.Start(ctx); err != nil {
		l.Fatal("startup", "Server failed", "HTTP_INTERNAL_ERROR", map[string]interface{}{"error": err.Error()})
	}
}
