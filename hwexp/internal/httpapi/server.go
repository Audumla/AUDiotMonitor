package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"hwexp/internal/config"
	"hwexp/internal/engine"
	"hwexp/internal/model"
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

// HandleDebugState returns a single consolidated JSON view of every device
// and its current measurements, joined on stable_id. This is the "everything
// in one place" endpoint.
func (s *Server) HandleDebugState(w http.ResponseWriter, r *http.Request) {
	devices := s.store.GetDevices()
	measurements := s.store.GetAllNormalized()

	// Index measurements by device ID
	byDevice := make(map[string][]model.NormalizedMeasurement, len(devices))
	for _, m := range measurements {
		byDevice[m.StableDeviceID] = append(byDevice[m.StableDeviceID], m)
	}

	type MeasurementView struct {
		LogicalName  string            `json:"logical_name"`
		MetricFamily string            `json:"metric_family"`
		Value        float64           `json:"value"`
		Unit         string            `json:"unit"`
		Labels       map[string]string `json:"labels,omitempty"`
		Quality      string            `json:"quality"`
		Timestamp    time.Time         `json:"timestamp"`
	}

	type DeviceView struct {
		StableID          string                 `json:"stable_id"`
		DisplayName       string                 `json:"display_name"`
		DeviceClass       string                 `json:"device_class"`
		DeviceSubclass    string                 `json:"device_subclass,omitempty"`
		Vendor            string                 `json:"vendor,omitempty"`
		Model             string                 `json:"model,omitempty"`
		Driver            string                 `json:"driver,omitempty"`
		Bus               string                 `json:"bus,omitempty"`
		Location          string                 `json:"location,omitempty"`
		Source            string                 `json:"source"`
		Capabilities      []string               `json:"capabilities"`
		RawIdentifiers    map[string]string      `json:"raw_identifiers,omitempty"`
		AdapterMetadata   map[string]interface{} `json:"adapter_metadata,omitempty"`
		Measurements      []MeasurementView      `json:"measurements"`
	}

	views := make([]DeviceView, 0, len(devices))
	for _, dev := range devices {
		ms := byDevice[dev.StableID]
		mvs := make([]MeasurementView, 0, len(ms))
		for _, m := range ms {
			mvs = append(mvs, MeasurementView{
				LogicalName:  m.LogicalName,
				MetricFamily: m.MetricFamily,
				Value:        m.Value,
				Unit:         m.Unit,
				Labels:       m.Labels,
				Quality:      m.Quality,
				Timestamp:    m.Timestamp,
			})
		}
		views = append(views, DeviceView{
			StableID:        dev.StableID,
			DisplayName:     dev.DisplayName,
			DeviceClass:     dev.DeviceClass,
			DeviceSubclass:  dev.DeviceSubclass,
			Vendor:          dev.Vendor,
			Model:           dev.Model,
			Driver:          dev.Driver,
			Bus:             dev.Bus,
			Location:        dev.Location,
			Source:          dev.Source,
			Capabilities:    dev.Capabilities,
			RawIdentifiers:  dev.RawIdentifiers,
			AdapterMetadata: dev.AdapterMetadata,
			Measurements:    mvs,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"schema_version": "1.0.0",
		"generated_at":   time.Now().UTC().Format(time.RFC3339),
		"host":           s.cfg.Identity.Host,
		"device_count":   len(views),
		"devices":        views,
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
	devices := s.store.GetDevices()
	sm := s.engine.GetSelfMetrics()
	
	// Group by MetricFamily for OpenMetrics formatting
	grouped := make(map[string][]string)
	types := make(map[string]string)

	for _, m := range measurements {
		types[m.MetricFamily] = m.MetricType

		// Build labels string: {host="x",sensor="y"}
		var lb strings.Builder
		lb.WriteByte('{')
		lb.WriteString(`host="`)
		lb.WriteString(s.cfg.Identity.Host)
		lb.WriteByte('"')
		for k, v := range m.Labels {
			if k == "host" {
				continue // host is always injected from config above
			}
			lb.WriteByte(',')
			lb.WriteString(k)
			lb.WriteString(`="`)
			lb.WriteString(v)
			lb.WriteByte('"')
		}
		lb.WriteByte('}')

		line := fmt.Sprintf("%s%s %f", m.MetricFamily, lb.String(), m.Value)
		grouped[m.MetricFamily] = append(grouped[m.MetricFamily], line)
	}

	// Write output
	fmt.Fprintf(w, "# HELP hwexp_up Exporter is running\n# TYPE hwexp_up gauge\nhwexp_up{host=%q} 1\n", s.cfg.Identity.Host)
	
	// Device Info Metrics
	fmt.Fprintf(w, "# HELP hw_device_info Metadata about discovered hardware devices\n# TYPE hw_device_info gauge\n")
	for _, dev := range devices {
		var lb strings.Builder
		lb.WriteByte('{')
		lb.WriteString(fmt.Sprintf(`host=%q`, s.cfg.Identity.Host))
		lb.WriteString(fmt.Sprintf(`,device_id=%q`, dev.StableID))
		lb.WriteString(fmt.Sprintf(`,device_class=%q`, dev.DeviceClass))
		lb.WriteString(fmt.Sprintf(`,vendor=%q`, dev.Vendor))
		lb.WriteString(fmt.Sprintf(`,model=%q`, dev.Model))
		lb.WriteString(fmt.Sprintf(`,driver=%q`, dev.Driver))
		
		// Add some metadata if available
		if bios, ok := dev.AdapterMetadata["bios_version"].(string); ok {
			lb.WriteString(fmt.Sprintf(`,bios_version=%q`, bios))
		}
		if cpuCores, ok := dev.AdapterMetadata["cores"].(int); ok {
			lb.WriteString(fmt.Sprintf(`,cpu_cores=%q`, strconv.Itoa(cpuCores)))
		}
		if cpuThreads, ok := dev.AdapterMetadata["threads"].(int); ok {
			lb.WriteString(fmt.Sprintf(`,cpu_threads=%q`, strconv.Itoa(cpuThreads)))
		}

		lb.WriteByte('}')
		fmt.Fprintf(w, "hw_device_info%s 1\n", lb.String())
	}

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

func (s *Server) Start(ctx context.Context) error {
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

	stateHandler := s.HandleDebugState
	if s.cfg.Security.AuthMode == "bearer_token" {
		stateHandler = s.authStore.Middleware(stateHandler, "debug:read")
	}

	mux.HandleFunc("/metrics", metricsHandler)
	mux.HandleFunc("/debug/state", stateHandler)
	mux.HandleFunc("/debug/discovery", discoveryHandler)
	mux.HandleFunc("/debug/catalog", catalogHandler)
	mux.HandleFunc("/debug/raw", rawHandler)
	mux.HandleFunc("/debug/mappings", mappingsHandler)

	srv := &http.Server{
		Addr:    s.cfg.Server.ListenAddress,
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}

