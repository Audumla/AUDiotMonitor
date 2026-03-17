package store

import (
	"hwexp/internal/model"
	"sync"
)

// StateStore holds the concurrency-safe latest snapshot of the system.
// As per spec Section 6: "All stores MUST be atomically replaceable per refresh cycle."
type StateStore struct {
	mu           sync.RWMutex
	devices      map[string]model.DiscoveredDevice
	measurements map[string]model.NormalizedMeasurement
	raw          map[string]model.RawMeasurement
	decisions    []model.MappingDecision
}

func NewStateStore() *StateStore {
	return &StateStore{
		devices:      make(map[string]model.DiscoveredDevice),
		measurements: make(map[string]model.NormalizedMeasurement),
		raw:          make(map[string]model.RawMeasurement),
		decisions:    make([]model.MappingDecision, 0),
	}
}

// UpdateSnapshot atomically replaces the metrics maps.
func (s *StateStore) UpdateSnapshot(
	newDevices map[string]model.DiscoveredDevice,
	newMeasurements map[string]model.NormalizedMeasurement,
	newRaw map[string]model.RawMeasurement,
	newDecisions []model.MappingDecision,
) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices = newDevices
	s.measurements = newMeasurements
	s.raw = newRaw
	s.decisions = newDecisions
}

// GetAllNormalized returns a copy of all current normalized measurements.
func (s *StateStore) GetAllNormalized() []model.NormalizedMeasurement {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	res := make([]model.NormalizedMeasurement, 0, len(s.measurements))
	for _, m := range s.measurements {
		res = append(res, m)
	}
	return res
}

func (s *StateStore) GetDevices() []model.DiscoveredDevice {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]model.DiscoveredDevice, 0, len(s.devices))
	for _, d := range s.devices {
		res = append(res, d)
	}
	return res
}

func (s *StateStore) GetRaw() []model.RawMeasurement {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]model.RawMeasurement, 0, len(s.raw))
	for _, r := range s.raw {
		res = append(res, r)
	}
	return res
}

func (s *StateStore) GetDecisions() []model.MappingDecision {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.decisions
}
