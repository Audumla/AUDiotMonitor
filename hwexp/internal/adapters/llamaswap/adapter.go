package llamaswap

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"hwexp/internal/model"
)

type Adapter struct {
	endpoint string
	client   *http.Server
}

func NewAdapter(endpoint string) *Adapter {
	if endpoint == "" {
		endpoint = "http://localhost:50099"
	}
	return &Adapter{
		endpoint: endpoint,
	}
}

type ModelResponse struct {
	Data []struct {
		ID      string `json:"id"`
		OwnedBy string `json:"owned_by"`
	} `json:"data"`
}

func (a *Adapter) Discover(ctx context.Context) ([]model.DiscoveredDevice, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", a.endpoint+"/v1/models", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var mResp ModelResponse
	if err := json.NewDecoder(resp.Body).Decode(&mResp); err != nil {
		return nil, err
	}

	var devices []model.DiscoveredDevice
	now := time.Now().UTC()
	for _, m := range mResp.Data {
		devices = append(devices, model.DiscoveredDevice{
			StableID:     "llm-" + m.ID,
			Platform:     "software",
			Source:       "llamaswap",
			DeviceClass:  "llm",
			Vendor:       m.OwnedBy,
			Model:        m.ID,
			DisplayName:  "LLM: " + m.ID,
			Capabilities: []string{"inference"},
			Present:      true,
			FirstSeen:    now,
			LastSeen:     now,
		})
	}

	return devices, nil
}

func (a *Adapter) Poll(ctx context.Context) ([]model.RawMeasurement, error) {
	devices, err := a.Discover(ctx)
	if err != nil {
		return nil, err
	}

	var all []model.RawMeasurement
	now := time.Now().UTC()

	// In a real scenario, we might query a /health or /metrics endpoint on llamaswap
	// for specific performance data. For now, we report model status.
	for _, dev := range devices {
		all = append(all, model.RawMeasurement{
			MeasurementID:  "llamaswap:" + dev.StableID + ":active",
			StableDeviceID: dev.StableID,
			Source:         "llamaswap",
			RawName:        "model_active",
			RawValue:       1.0,
			RawUnit:        "count",
			Timestamp:      now,
			Quality:        "good",
			ComponentHint:  "inference",
			SensorHint:     "status",
		})
	}

	return all, nil
}
