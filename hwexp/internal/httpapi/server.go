package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"hwexp/internal/config"
	"hwexp/internal/store"
)

type Server struct {
	cfg       *config.Config
	store     *store.StateStore
	authStore *AuthStore
}

func NewServer(cfg *config.Config, s *store.StateStore, auth *AuthStore) *Server {
	return &Server{
		cfg:       cfg,
		store:     s,
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
	fmt.Fprintf(w, "# HELP hwexp_up Test harness is running\n# TYPE hwexp_up gauge\nhwexp_up 1\n")
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
	mux.HandleFunc("/version", s.HandleVersion)

	// Potentially protected endpoints
	metricsHandler := s.HandleMetrics
	discoveryHandler := s.HandleDebugDiscovery
	rawHandler := s.HandleDebugRaw
	mappingsHandler := s.HandleDebugMappings

	if s.cfg.Security.AuthMode == "bearer_token" {
		discoveryHandler = s.authStore.Middleware(discoveryHandler, "debug:read")
		rawHandler = s.authStore.Middleware(rawHandler, "debug:read")
		mappingsHandler = s.authStore.Middleware(mappingsHandler, "debug:read")
		// metrics might be protected too depending on policy, spec says "MAY remain open"
	}

	mux.HandleFunc("/metrics", metricsHandler)
	mux.HandleFunc("/debug/discovery", discoveryHandler)
	mux.HandleFunc("/debug/raw", rawHandler)
	mux.HandleFunc("/debug/mappings", mappingsHandler)

	return http.ListenAndServe(s.cfg.Server.ListenAddress, mux)
}

