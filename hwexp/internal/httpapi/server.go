package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"hwexp/internal/config"
	"hwexp/internal/engine"
	"hwexp/internal/store"
)

type Server struct {
	cfg       *config.Config
	store     *store.StateStore
	engine    *engine.Engine
	authStore *AuthStore
}

func NewServer(cfg *config.Config, s *store.StateStore, e *engine.Engine, auth *AuthStore) *Server {
	return &Server{
		cfg:       cfg,
		store:     s,
		engine:    e,
		authStore: auth,
	}
}

func (s *Server) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"schema_version": "1.0.0",
		"status":         "ok",
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) HandleReadyz(w http.ResponseWriter, r *http.Request) {
	if s.store.IsReady() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ready",
		})
	} else {
		http.Error(w, `{"status":"not ready"}`, http.StatusServiceUnavailable)
	}
}

func (s *Server) HandleVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"schema_version":        "1.0.0",
		"exporter_version":      "0.1.0",
		"api_schema_version":    "1.0.0",
		"metric_schema_version": "1.0.0",
		"config_schema_version": "1.0.0",
		"platform":              s.cfg.Identity.Platform,
	})
}

func (s *Server) HandleDebugDiscovery(w http.ResponseWriter, r *http.Request) {
	devices := s.store.GetDevices()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"schema_version": "1.0.0",
		"generated_at":   time.Now().UTC().Format(time.RFC3339),
		"host":           s.cfg.Identity.Host,
		"devices":        devices,
		"summary": map[string]interface{}{
			"device_count":         len(devices),
			"unknown_device_count": 0,
		},
	})
}

func (s *Server) HandleDebugCatalog(w http.ResponseWriter, r *http.Request) {
	measurements := s.store.GetAllNormalized()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"schema_version": "1.0.0",
		"generated_at":   time.Now().UTC().Format(time.RFC3339),
		"host":           s.cfg.Identity.Host,
		"items":          measurements,
	})
}

func (s *Server) HandleDebugRaw(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.Debug.EnableRawEndpoint {
		s.authStore.jsonError(w, http.StatusForbidden, "HTTP_FORBIDDEN", "Raw debug endpoint is disabled")
		return
	}
	measurements := s.store.GetRaw()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"schema_version": "1.0.0",
		"generated_at":   time.Now().UTC().Format(time.RFC3339),
		"items":          measurements,
	})
}

func (s *Server) HandleDebugMappings(w http.ResponseWriter, r *http.Request) {
	decisions := s.store.GetDecisions()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"schema_version": "1.0.0",
		"generated_at":   time.Now().UTC().Format(time.RFC3339),
		"items":          decisions,
	})
}

func (s *Server) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	
	measurements := s.store.GetAllNormalized()
	sm := s.engine.GetSelfMetrics()
	
	// Group by MetricFamily for OpenMetrics formatting
	grouped := make(map[string][]string)
	types := make(map[string]string)

	for _, m := range measurements {
		types[m.MetricFamily] = m.MetricType
		
		// Build labels string: {host="x",sensor="y"}
		labelStr := ""
		ls := []string{
			fmt.Sprintf(`host="%s"`, s.cfg.Identity.Host),
		}
		for k, v := range m.Labels {
			if k == "host" {
				continue // host is always injected from config above
			}
			ls = append(ls, fmt.Sprintf(`%s="%s"`, k, v))
		}
		labelStr = "{"
		for i, l := range ls {
			if i > 0 { labelStr += "," }
			labelStr += l
		}
		labelStr += "}"
		
		line := fmt.Sprintf("%s%s %f", m.MetricFamily, labelStr, m.Value)
		grouped[m.MetricFamily] = append(grouped[m.MetricFamily], line)
	}

	// Write output
	fmt.Fprintf(w, "# HELP hwexp_up Exporter is running\n# TYPE hwexp_up gauge\nhwexp_up{host=%q} 1\n", s.cfg.Identity.Host)
	
	// Self-metrics
	fmt.Fprintf(w, "# HELP hwexp_adapter_refresh_duration_seconds Duration of the last adapter refresh\n# TYPE hwexp_adapter_refresh_duration_seconds gauge\nhwexp_adapter_refresh_duration_seconds{host=%q} %f\n", s.cfg.Identity.Host, sm.LastRefreshDuration.Seconds())
	
	valSuccess := 0.0
	if sm.LastRefreshSuccess { valSuccess = 1.0 }
	fmt.Fprintf(w, "# HELP hwexp_adapter_refresh_success Whether the last adapter refresh was successful\n# TYPE hwexp_adapter_refresh_success gauge\nhwexp_adapter_refresh_success{host=%q} %f\n", s.cfg.Identity.Host, valSuccess)
	
	fmt.Fprintf(w, "# HELP hwexp_adapter_last_success_unixtime Last successful refresh timestamp\n# TYPE hwexp_adapter_last_success_unixtime gauge\nhwexp_adapter_last_success_unixtime{host=%q} %d\n", s.cfg.Identity.Host, sm.LastSuccessTime.Unix())
	
	fmt.Fprintf(w, "# HELP hwexp_discovered_devices Number of discovered devices\n# TYPE hwexp_discovered_devices gauge\nhwexp_discovered_devices{host=%q} %d\n", s.cfg.Identity.Host, sm.DiscoveredDevices)
	
	fmt.Fprintf(w, "# HELP hwexp_mapping_failures_total Total number of mapping failures\n# TYPE hwexp_mapping_failures_total counter\nhwexp_mapping_failures_total{host=%q} %d\n", s.cfg.Identity.Host, sm.MappingFailures)

	for family, lines := range grouped {
		fmt.Fprintf(w, "# TYPE %s %s\n", family, types[family])
		for _, line := range lines {
			fmt.Fprintf(w, "%s\n", line)
		}
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()
	
	// Open endpoints
	mux.HandleFunc("/healthz", s.HandleHealthz)
	mux.HandleFunc("/readyz", s.HandleReadyz)
	mux.HandleFunc("/version", s.HandleVersion)

	// Potentially protected endpoints
	metricsHandler := s.HandleMetrics
	discoveryHandler := s.HandleDebugDiscovery
	catalogHandler := s.HandleDebugCatalog
	rawHandler := s.HandleDebugRaw
	mappingsHandler := s.HandleDebugMappings

	if s.cfg.Security.AuthMode == "bearer_token" {
		discoveryHandler = s.authStore.Middleware(discoveryHandler, "debug:read")
		catalogHandler = s.authStore.Middleware(catalogHandler, "debug:read")
		rawHandler = s.authStore.Middleware(rawHandler, "debug:read")
		mappingsHandler = s.authStore.Middleware(mappingsHandler, "debug:read")
		// metrics might be protected too depending on policy, spec says "MAY remain open"
	}

	mux.HandleFunc("/metrics", metricsHandler)
	mux.HandleFunc("/debug/discovery", discoveryHandler)
	mux.HandleFunc("/debug/catalog", catalogHandler)
	mux.HandleFunc("/debug/raw", rawHandler)
	mux.HandleFunc("/debug/mappings", mappingsHandler)

	return http.ListenAndServe(s.cfg.Server.ListenAddress, mux)
}

