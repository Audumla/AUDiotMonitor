package main

import (
	"context"
	"flag"
	"log"

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
	fixturePath := flag.String("fixture", "tests/fixtures/sample_hwmon.json", "Path to mock fixture JSON")
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

	// 5. Initialize Mock Adapter
	adapter := mock.NewAdapter(*fixturePath)
	if err := adapter.Load(); err != nil {
		l.Fatal("startup", "Failed to load fixture", "ADAPTER_INIT_FAILED", map[string]interface{}{"error": err.Error()})
	}

	// 6. Initialize State Store
	stateStore := store.NewStateStore()

	// 7. Initialize Mapping Engine
	mapperEngine, err := mapper.NewEngine(rules)
	if err != nil {
		l.Fatal("startup", "Failed to initialize mapper", "INTERNAL_ERROR", map[string]interface{}{"error": err.Error()})
	}

	// 8. Initialize and start Core Engine loop
	coreEngine := engine.NewEngine(
		stateStore,
		mapperEngine,
		[]engine.Adapter{adapter},
	)
	
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
