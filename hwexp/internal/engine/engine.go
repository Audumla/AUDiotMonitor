package engine

import (
	"context"
	"log"
	"time"

	"hwexp/internal/model"
	"hwexp/internal/mapper"
	"hwexp/internal/store"
)

// Adapter describes the interface required to pull data from hardware.
type Adapter interface {
	Discover(ctx context.Context) ([]model.DiscoveredDevice, error)
	Poll(ctx context.Context) ([]model.RawMeasurement, error)
}

type Engine struct {
	store        *store.StateStore
	mapper       *mapper.Engine
	adapters     []Adapter
	pollInterval time.Duration
}

func NewEngine(s *store.StateStore, m *mapper.Engine, adapters []Adapter) *Engine {
	return &Engine{
		store:        s,
		mapper:       m,
		adapters:     adapters,
		pollInterval: 5 * time.Second, // Default from spec
	}
}

func (e *Engine) Start(ctx context.Context) {
	go e.loop(ctx)
}

func (e *Engine) loop(ctx context.Context) {
	ticker := time.NewTicker(e.pollInterval)
	defer ticker.Stop()

	// Initial synchronous poll so we are ready
	e.executeCycle(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.executeCycle(ctx)
		}
	}
}

func (e *Engine) executeCycle(ctx context.Context) {
	newDevices := make(map[string]model.DiscoveredDevice)
	newMeasurements := make(map[string]model.NormalizedMeasurement)
	newRaw := make(map[string]model.RawMeasurement)
	var newDecisions []model.MappingDecision

	for _, a := range e.adapters {
		// 1. Discover
		devices, err := a.Discover(ctx)
		if err != nil {
			log.Printf("Adapter discovery error: %v", err)
			continue
		}
		for _, d := range devices {
			newDevices[d.StableID] = d
		}

		// 2. Poll
		raws, err := a.Poll(ctx)
		if err != nil {
			log.Printf("Adapter poll error: %v", err)
			continue
		}

		// 3. Normalize
		for _, r := range raws {
			newRaw[r.MeasurementID] = r

			device, exists := newDevices[r.StableDeviceID]
			if !exists {
				// We have a measurement for a device we didn't discover this cycle.
				// In a real app, we'd check grace periods here.
				continue 
			}

			norm, decision := e.mapper.Map(device, r)
			if decision != nil {
				newDecisions = append(newDecisions, *decision)
			}
			
			if norm != nil {
				// Combine logical name and device for uniqueness
				key := norm.StableDeviceID + ":" + norm.LogicalName
				newMeasurements[key] = *norm
			}
		}
	}

	// 4. Update snapshot atomically
	e.store.UpdateSnapshot(newDevices, newMeasurements, newRaw, newDecisions)
}
