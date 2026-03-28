package gateway_manifest

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"hwexp/internal/model"
)

type Adapter struct {
	projectDir string
	localDir   string
	client     *http.Client

	mu        sync.RWMutex
	manifests []model.Manifest
	once      sync.Once
}

func NewAdapter(projectDir, localDir string) *Adapter {
	return &Adapter{
		projectDir: projectDir,
		localDir:   localDir,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (a *Adapter) startRefresher(ctx context.Context) {
	a.refresh()
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.refresh()
			}
		}
	}()
}

func (a *Adapter) refresh() {
	m, err := LoadManifests(a.projectDir, a.localDir)
	if err != nil {
		log.Printf("gateway_manifest: failed to refresh manifests: %v", err)
		return
	}
	a.mu.Lock()
	a.manifests = m
	a.mu.Unlock()
}

func (a *Adapter) Discover(ctx context.Context) ([]model.DiscoveredDevice, error) {
	a.once.Do(func() { a.startRefresher(ctx) })

	a.mu.RLock()
	manifests := a.manifests
	a.mu.RUnlock()

	var devices []model.DiscoveredDevice
	now := time.Now().UTC()

	for _, m := range manifests {
		devices = append(devices, model.DiscoveredDevice{
			StableID:          "gateway-" + m.ID,
			Platform:          "software",
			Source:            "gateway_manifest",
			DeviceClass:       "gateway_component",
			Vendor:            "audia",
			Model:             m.ID,
			DisplayName:       m.DisplayName,
			LogicalDeviceName: m.ID,
			Capabilities:      []string{"health", "metrics"},
			Present:           true,
			FirstSeen:         now,
			LastSeen:          now,
		})
	}
	return devices, nil
}

func (a *Adapter) Poll(ctx context.Context) ([]model.RawMeasurement, error) {
	a.once.Do(func() { a.startRefresher(ctx) })

	a.mu.RLock()
	manifests := a.manifests
	a.mu.RUnlock()

	var allMetrics []model.RawMeasurement
	now := time.Now().UTC()

	for _, m := range manifests {
		baseURL := fmt.Sprintf("http://%s:%d", m.Connection.Host, m.Connection.Port)
		stableID := "gateway-" + m.ID

		// 1. Health Check (Only for HTTP based components or generic)
		up := 0.0
		healthURL := baseURL + m.Health.Endpoint
		// Default to up for non-http if no health check endpoint?
		// Actually, even exec/file components might have an HTTP control plane.
		// If endpoint starts with /, it's HTTP.
		if strings.HasPrefix(m.Health.Endpoint, "/") {
			hClient := a.client
			if m.Health.TimeoutS > 0 {
				hClient = &http.Client{Timeout: time.Duration(m.Health.TimeoutS) * time.Second}
			}
			req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, healthURL, nil)
			if reqErr == nil {
				if m.Connection.Auth.Type == "bearer" && m.Connection.Auth.TokenEnv != "" {
					if token := os.Getenv(m.Connection.Auth.TokenEnv); token != "" {
						req.Header.Set("Authorization", "Bearer "+token)
					}
				}
			}
			var resp *http.Response
			var err error
			if reqErr == nil {
				resp, err = hClient.Do(req)
			} else {
				err = reqErr
			}
			if err == nil {
				if resp.StatusCode == m.Health.ExpectStatus {
					up = 1.0
				}
				resp.Body.Close()
			}
		} else {
			up = 1.0 // Fallback
		}

		allMetrics = append(allMetrics, model.RawMeasurement{
			MeasurementID:  fmt.Sprintf("gateway:%s:up", m.ID),
			StableDeviceID: stableID,
			Source:         "gateway_manifest",
			RawName:        "component_up",
			RawValue:       up,
			Timestamp:      now,
			Quality:        "good",
		})

		// 2. Info Metric
		allMetrics = append(allMetrics, model.RawMeasurement{
			MeasurementID:  fmt.Sprintf("gateway:%s:info", m.ID),
			StableDeviceID: stableID,
			Source:         "gateway_manifest",
			RawName:        "component_info",
			RawValue:       1.0,
			Timestamp:      now,
			Quality:        "good",
		})

		// Correlation metadata helper
		injectCorrelation := func(rm *model.RawMeasurement) {
			if m.Correlation != nil {
				if rm.Metadata == nil {
					rm.Metadata = make(map[string]string)
				}
				rm.Metadata["correlation_pci_slot"] = m.Correlation.PCISlot
				rm.Metadata["correlation_class"] = m.Correlation.DeviceClass
			}
		}

		// 3. Optional Two-Tier Discovery (e.g. llama-swap)
		if up == 1.0 && m.Discovery != nil {
			discoveryURL := baseURL + m.Discovery.Endpoint
			dResp, err := a.client.Get(discoveryURL)
			if err == nil {
				body, _ := io.ReadAll(dResp.Body)
				dResp.Body.Close()

				var data struct {
					Data []map[string]interface{} `json:"data"`
				}
				if err := json.Unmarshal(body, &data); err == nil {
					aggregates := make(map[string]float64)
					for _, instance := range data.Data {
						id, _ := instance["id"].(string)
						active := false
						if val, ok := instance[m.Discovery.ActivityField].(float64); ok && val > 0 {
							active = true
						}

						if active {
							port := 0
							if p, ok := instance[m.Discovery.BackendPortField].(float64); ok {
								port = int(p)
							}

							if port > 0 {
								instanceMetrics := a.scrapeInstance(m.Connection.Host, port, stableID, id, now)
								for _, im := range instanceMetrics {
									injectCorrelation(&im)
									allMetrics = append(allMetrics, im)
									if im.RawName == "gateway_llamacpp_prompt_tokens_total" {
										aggregates[im.RawName] += im.RawValue
									}
								}
							}
						}
					}
					for name, val := range aggregates {
						rm := model.RawMeasurement{
							MeasurementID:  fmt.Sprintf("gateway:%s:aggregate:%s", m.ID, name),
							StableDeviceID: stableID,
							Source:         "gateway_manifest",
							RawName:        name,
							RawValue:       val,
							Timestamp:      now,
							Quality:        "good",
							Metadata: map[string]string{
								"model_name": "active",
							},
						}
						injectCorrelation(&rm)
						allMetrics = append(allMetrics, rm)
					}
				}
			}
		}

		// 4. Standard Metrics Scrape
		if up == 1.0 {
			for _, mc := range m.Metrics {
				var body []byte
				var err error

				sourceType := mc.SourceType
				if sourceType == "" {
					sourceType = "http"
				}

				switch sourceType {
				case "http":
					metricURL := baseURL + mc.Endpoint
					req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, metricURL, nil)
					if reqErr != nil {
						err = reqErr
						break
					}
					if m.Connection.Auth.Type == "bearer" && m.Connection.Auth.TokenEnv != "" {
						if token := os.Getenv(m.Connection.Auth.TokenEnv); token != "" {
							req.Header.Set("Authorization", "Bearer "+token)
						}
					}
					mResp, err2 := a.client.Do(req)
					if err2 != nil {
						err = err2
					} else {
						body, _ = io.ReadAll(mResp.Body)
						mResp.Body.Close()
					}
				case "exec":
					cmdParts := strings.Fields(mc.Endpoint)
					if len(cmdParts) > 0 {
						c := exec.CommandContext(ctx, cmdParts[0], cmdParts[1:]...)
						body, err = c.Output()
					}
				case "file":
					body, err = os.ReadFile(mc.Endpoint)
				}

				if err != nil {
					continue
				}

				var val float64
				if mc.SourceFormat == "prometheus" {
					val, err = ExtractPrometheusValue(body, mc.Extract)
				} else {
					val, err = ExtractJSONValue(body, mc.Extract)
				}

				if err != nil {
					log.Printf("gateway_manifest: component %s extraction failed for %s: %v", m.ID, mc.ID, err)
					continue
				}

				rm := model.RawMeasurement{
					MeasurementID:  fmt.Sprintf("gateway:%s:%s", m.ID, mc.ID),
					StableDeviceID: stableID,
					Source:         "gateway_manifest",
					RawName:        mc.PrometheusName,
					RawValue:       val,
					Timestamp:      now,
					Quality:        "good",
				}
				injectCorrelation(&rm)
				allMetrics = append(allMetrics, rm)
			}
		}
	}

	return allMetrics, nil
}

func (a *Adapter) scrapeInstance(host string, port int, stableID, modelID string, now time.Time) []model.RawMeasurement {
	url := fmt.Sprintf("http://%s:%d/metrics", host, port)
	resp, err := a.client.Get(url)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	// Extract all metrics from the backend (Prometheus format)
	// We'll use a simpler version of ExtractPrometheusValue that returns all metrics
	return parseAllPrometheusMetrics(body, stableID, modelID, now)
}

func parseAllPrometheusMetrics(data []byte, stableID, modelID string, now time.Time) []model.RawMeasurement {
	var results []model.RawMeasurement
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		name := fields[0]
		// remove labels for RawName, they will be added later or we can inject model_name label here via metadata
		if idx := strings.Index(name, "{"); idx > 0 {
			name = name[:idx]
		}

		val, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			continue
		}

		results = append(results, model.RawMeasurement{
			MeasurementID:  fmt.Sprintf("gateway:instance:%s:%s", modelID, name),
			StableDeviceID: stableID,
			Source:         "gateway_manifest",
			RawName:        name,
			RawValue:       val,
			Timestamp:      now,
			Quality:        "good",
			Metadata: map[string]string{
				"model_name": modelID,
			},
		})
	}
	return results
}
