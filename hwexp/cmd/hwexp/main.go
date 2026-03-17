package main

import (
	"context"
	"flag"
	"log"

	"hwexp/internal/adapters/hwmon"
	"hwexp/internal/adapters/mock"
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

	// 3. Load Mapping Rules
	rules, err := mapper.LoadRules(cfg.Mapping.RulesFile)
	if err != nil {
		l.Fatal("startup", "Failed to load rules", "CFG_LOAD_FAILED", map[string]interface{}{"error": err.Error()})
	}
	l.Info("startup", "Loaded mapping rules", map[string]interface{}{"count": len(rules), "file": cfg.Mapping.RulesFile})

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
	} else {
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l.Info("startup", "Starting background telemetry loop", map[string]interface{}{"interval": cfg.Server.RefreshInterval})
	coreEngine.Start(ctx)

	// 9. Start HTTP Server
	server := httpapi.NewServer(cfg, stateStore, authStore)
	l.Info("startup", "HTTP server listening", map[string]interface{}{"address": cfg.Server.ListenAddress})
	if err := server.Start(); err != nil {
		l.Fatal("startup", "Server failed", "HTTP_INTERNAL_ERROR", map[string]interface{}{"error": err.Error()})
	}
}
