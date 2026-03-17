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
	// Update timestamps for realism in the test harness
	now := time.Now()
	for i := range a.fixture.Devices {
		a.fixture.Devices[i].LastSeen = now
	}
	return a.fixture.Devices, nil
}

func (a *Adapter) Poll(ctx context.Context) ([]model.RawMeasurement, error) {
	if a.fixture == nil {
		return nil, nil
	}
	now := time.Now()
	for i := range a.fixture.Measurements {
		a.fixture.Measurements[i].Timestamp = now
	}
	return a.fixture.Measurements, nil
}
