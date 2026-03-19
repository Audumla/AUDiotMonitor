package mock

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"hwexp/internal/model"
)

type Fixture struct {
	Devices      []model.DiscoveredDevice `json:"devices"`
	Measurements []model.RawMeasurement   `json:"measurements"`
}

type Adapter struct {
	fixturePath string
	fixture     *Fixture
}

func NewAdapter(fixturePath string) *Adapter {
	return &Adapter{fixturePath: fixturePath}
}

func (a *Adapter) Load() error {
	data, err := os.ReadFile(a.fixturePath)
	if err != nil {
		return err
	}
	var f Fixture
	if err := json.Unmarshal(data, &f); err != nil {
		return err
	}
	a.fixture = &f
	return nil
}

func (a *Adapter) Discover(ctx context.Context) ([]model.DiscoveredDevice, error) {
	if a.fixture == nil {
		return nil, nil
	}
	now := time.Now()
	devices := make([]model.DiscoveredDevice, len(a.fixture.Devices))
	copy(devices, a.fixture.Devices)
	for i := range devices {
		devices[i].LastSeen = now
	}
	return devices, nil
}

func (a *Adapter) Poll(ctx context.Context) ([]model.RawMeasurement, error) {
	if a.fixture == nil {
		return nil, nil
	}
	now := time.Now()
	measurements := make([]model.RawMeasurement, len(a.fixture.Measurements))
	copy(measurements, a.fixture.Measurements)
	for i := range measurements {
		measurements[i].Timestamp = now
	}
	return measurements, nil
}
