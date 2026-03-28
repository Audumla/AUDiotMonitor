package linux_system

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"hwexp/internal/model"
)

type Adapter struct {
	lastInterrupts map[string]uint64
	lastPoll       time.Time
}

func NewAdapter() *Adapter {
	return &Adapter{
		lastInterrupts: make(map[string]uint64),
	}
}

func (a *Adapter) Discover(ctx context.Context) ([]model.DiscoveredDevice, error) {
	now := time.Now().UTC()
	return []model.DiscoveredDevice{
		{
			StableID:     "system-host",
			Platform:     "linux",
			Source:       "linux_system",
			DeviceClass:  "motherboard",
			Capabilities: []string{"acpi", "interrupts", "cpuidle"},
			Present:      true,
			FirstSeen:    now,
			LastSeen:     now,
		},
	}, nil
}

func (a *Adapter) Poll(ctx context.Context) ([]model.RawMeasurement, error) {
	var metrics []model.RawMeasurement
	now := time.Now().UTC()

	// 1. ACPI Thermal Zones
	zones, _ := filepath.Glob("/sys/class/thermal/thermal_zone*/temp")
	for _, z := range zones {
		id := filepath.Base(filepath.Dir(z))
		data, err := os.ReadFile(z)
		if err == nil {
			val, _ := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
			metrics = append(metrics, model.RawMeasurement{
				MeasurementID:  fmt.Sprintf("system:thermal:%s", id),
				StableDeviceID: "system-host",
				Source:         "linux_system",
				RawName:        "thermal_zone_temp",
				RawValue:       val,
				Timestamp:      now,
				Quality:        "good",
				Metadata:       map[string]string{"zone": id},
			})
		}
	}

	// 2. CPU C-State Residency
	// We'll just take the average across all CPUs for now to avoid metric explosion,
	// or just provide core 0 as a representative sample.
	cstates, _ := filepath.Glob("/sys/devices/system/cpu/cpu0/cpuidle/state*/residency")
	for _, cs := range cstates {
		stateID := filepath.Base(filepath.Dir(cs))
		data, err := os.ReadFile(cs)
		if err == nil {
			val, _ := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
			metrics = append(metrics, model.RawMeasurement{
				MeasurementID:  fmt.Sprintf("system:cpuidle:%s", stateID),
				StableDeviceID: "system-host",
				Source:         "linux_system",
				RawName:        "cpu_cstate_residency",
				RawValue:       val,
				Timestamp:      now,
				Quality:        "good",
				Metadata:       map[string]string{"state": stateID},
			})
		}
	}

	// 3. Interrupts (Delta)
	currInterrupts := a.readInterrupts()
	if !a.lastPoll.IsZero() {
		duration := now.Sub(a.lastPoll).Seconds()
		totalDelta := uint64(0)
		for k, v := range currInterrupts {
			if last, ok := a.lastInterrupts[k]; ok && v >= last {
				totalDelta += (v - last)
			}
		}
		metrics = append(metrics, model.RawMeasurement{
			MeasurementID:  "system:interrupts",
			StableDeviceID: "system-host",
			Source:         "linux_system",
			RawName:        "system_interrupts_rate",
			RawValue:       float64(totalDelta) / duration,
			Timestamp:      now,
			Quality:        "good",
		})
	}
	a.lastInterrupts = currInterrupts
	a.lastPoll = now

	// 4. EDAC (Memory ECC)
	edac, _ := filepath.Glob("/sys/devices/system/edac/mc/mc*/ce_count")
	for _, e := range edac {
		id := filepath.Base(filepath.Dir(e))
		data, err := os.ReadFile(e)
		if err == nil {
			val, _ := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
			metrics = append(metrics, model.RawMeasurement{
				MeasurementID:  fmt.Sprintf("system:edac:ce:%s", id),
				StableDeviceID: "system-host",
				Source:         "linux_system",
				RawName:        "memory_ecc_ce_count",
				RawValue:       val,
				Timestamp:      now,
				Quality:        "good",
				Metadata:       map[string]string{"mc": id},
			})
		}
	}

	return metrics, nil
}

func (a *Adapter) readInterrupts() map[string]uint64 {
	res := make(map[string]uint64)
	f, err := os.Open("/proc/interrupts")
	if err != nil {
		return res
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		
		// First field is interrupt ID (e.g. "0:", "RES:")
		id := fields[0]
		if !strings.HasSuffix(id, ":") {
			continue
		}

		sum := uint64(0)
		for _, f := range fields[1:] {
			// Stop if we hit non-numeric (the interrupt description)
			val, err := strconv.ParseUint(f, 10, 64)
			if err != nil {
				break
			}
			sum += val
		}
		res[id] = sum
	}
	return res
}
