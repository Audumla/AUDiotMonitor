package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"hwexp/internal/adapters/hwmon"
	"hwexp/internal/adapters/linux_gpu"
	"hwexp/internal/adapters/linux_static"
	"hwexp/internal/adapters/llamaswap"
	"hwexp/internal/adapters/mock"
	"hwexp/internal/adapters/vendor_exec"
	"hwexp/internal/config"
	"hwexp/internal/engine"
	"hwexp/internal/httpapi"
	"hwexp/internal/logger"
	"hwexp/internal/mapper"
	"hwexp/internal/store"
)

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

	if *fixturePath == "" {
		l.Info("startup", "Using linux_static adapter", nil)
		adapters = append(adapters, linux_static.NewAdapter(""))
	}

	if cfg.Adapters.Llamaswap.Enabled && *fixturePath == "" {
		endpoint := "http://localhost:50099"
		if e, ok := cfg.Adapters.Llamaswap.Settings["endpoint"].(string); ok && e != "" {
			endpoint = e
		}
		l.Info("startup", "Using llamaswap adapter", map[string]interface{}{"endpoint": endpoint})
		adapters = append(adapters, llamaswap.NewAdapter(endpoint))
	}

	if cfg.Adapters.LinuxVendorExec.Enabled && *fixturePath == "" {
		scriptsDir := "/etc/hwexp/custom.d"
		if d, ok := cfg.Adapters.LinuxVendorExec.Settings["scripts_dir"].(string); ok && d != "" {
			scriptsDir = d
		}
		l.Info("startup", "Using vendor_exec adapter", map[string]interface{}{"dir": scriptsDir})
		adapters = append(adapters, vendor_exec.NewAdapter(scriptsDir))
	}

	if len(adapters) == 0 {
		l.Fatal("startup", "No adapters enabled in config and no --fixture provided", "CFG_LOAD_FAILED", nil)
	}

	// 6. Initialize State Store
	stateStore := store.NewStateStore()

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
