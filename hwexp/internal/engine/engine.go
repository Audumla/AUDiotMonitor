package engine

import (
	"context"
	"log"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
	"hwexp/internal/automapper"
	"hwexp/internal/mapper"
	"hwexp/internal/model"
	"hwexp/internal/store"
)

// Adapter describes the interface required to pull data from hardware.
type Adapter interface {
	Discover(ctx context.Context) ([]model.DiscoveredDevice, error)
	Poll(ctx context.Context) ([]model.RawMeasurement, error)
}

type SelfMetrics struct {
	LastRefreshDuration time.Duration
	LastRefreshSuccess  bool
	LastSuccessTime     time.Time
	DiscoveredDevices   int
	MappingFailures     int
}

type Engine struct {
	store          *store.StateStore
	mapper         *mapper.Engine
	adapters       []Adapter
	pollInterval   time.Duration
	autoMapEnabled bool
	generatedFile  string
	generatedRules map[string]model.MappingRule // keyed by rule ID
	mu             sync.RWMutex

	// Self-metrics
	metrics SelfMetrics
}

func NewEngine(s *store.StateStore, m *mapper.Engine, adapters []Adapter) *Engine {
	return &Engine{
		store:          s,
		mapper:         m,
		adapters:       adapters,
		pollInterval:   5 * time.Second,
		generatedRules: make(map[string]model.MappingRule),
	}
}

func (e *Engine) GetSelfMetrics() SelfMetrics {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.metrics
}

// EnableAutoMap turns on dynamic rule inference for unmapped measurements.
func (e *Engine) EnableAutoMap(generatedFile string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.autoMapEnabled = true
	e.generatedFile = generatedFile
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
	start := time.Now()
	newDevices := make(map[string]model.DiscoveredDevice)
	newMeasurements := make(map[string]model.NormalizedMeasurement)
	newRaw := make(map[string]model.RawMeasurement)
	var newDecisions []model.MappingDecision

	success := true
	mappingFailures := 0

	for _, a := range e.adapters {
		// 1. Discover
		devices, err := a.Discover(ctx)
		if err != nil {
			log.Printf("Adapter discovery error: %v", err)
			success = false
			continue
		}
		for _, d := range devices {
			newDevices[d.StableID] = d
		}

		// 2. Poll
		raws, err := a.Poll(ctx)
		if err != nil {
			log.Printf("Adapter poll error: %v", err)
			success = false
			continue
		}

		// 3. Normalize
		for _, r := range raws {
			newRaw[r.MeasurementID] = r

			device, exists := newDevices[r.StableDeviceID]
			if !exists {
				continue 
			}

			norm, decision := e.mapper.Map(device, r)
			if decision != nil {
				newDecisions = append(newDecisions, *decision)
				if decision.Decision == "ignored" {
					mappingFailures++
				}
			}

			// Auto-map: infer a rule for any measurement that has no manual mapping.
			e.mu.RLock()
			autoMapEnabled := e.autoMapEnabled
			e.mu.RUnlock()
			if autoMapEnabled && decision != nil && decision.Decision == "ignored" {
				if rule := automapper.InferRule(device, r); rule != nil {
					e.applyAutoRule(*rule)
				}
			}

			if norm != nil {
				key := norm.StableDeviceID + ":" + norm.LogicalName
				newMeasurements[key] = *norm
			}
		}
	}

	// 4. Update snapshot atomically
	e.store.UpdateSnapshot(newDevices, newMeasurements, newRaw, newDecisions)

	// 5. Update self-metrics
	e.mu.Lock()
	e.metrics.LastRefreshDuration = time.Since(start)
	e.metrics.LastRefreshSuccess = success
	if success {
		e.metrics.LastSuccessTime = time.Now()
	}
	e.metrics.DiscoveredDevices = len(newDevices)
	e.metrics.MappingFailures += mappingFailures
	e.mu.Unlock()
}

// applyAutoRule adds a newly inferred rule to the mapper and persists it.
func (e *Engine) applyAutoRule(rule model.MappingRule) {
	e.mu.Lock()
	_, alreadyKnown := e.generatedRules[rule.ID]
	if !alreadyKnown {
		e.generatedRules[rule.ID] = rule
	}
	e.mu.Unlock()

	if alreadyKnown {
		return
	}

	added, err := e.mapper.AddRules([]model.MappingRule{rule})
	if err != nil {
		log.Printf("automapper: failed to add rule %s: %v", rule.ID, err)
		return
	}
	if len(added) == 0 {
		return // duplicate, already in mapper from a previous load
	}

	log.Printf("automapper: inferred rule %s for %s/%s", rule.ID, rule.Match.DeviceClass, rule.Match.RawNameRegex)

	if e.generatedFile != "" {
		e.persistGeneratedRules()
	}
}

// persistGeneratedRules writes all auto-generated rules to the configured file.
func (e *Engine) persistGeneratedRules() {
	e.mu.Lock()
	rules := make([]model.MappingRule, 0, len(e.generatedRules))
	for _, r := range e.generatedRules {
		rules = append(rules, r)
	}
	e.mu.Unlock()

	wrapper := struct {
		SchemaVersion string             `yaml:"schema_version"`
		Rules         []model.MappingRule `yaml:"rules"`
	}{
		SchemaVersion: "1.0.0",
		Rules:         rules,
	}

	data, err := yaml.Marshal(wrapper)
	if err != nil {
		log.Printf("automapper: failed to marshal rules: %v", err)
		return
	}

	if err := os.WriteFile(e.generatedFile, data, 0644); err != nil {
		log.Printf("automapper: failed to write %s: %v", e.generatedFile, err)
	}
}
