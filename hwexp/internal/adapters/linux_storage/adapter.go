package linux_storage

import (
	"context"
	"encoding/json"
	"fmt"
	"hwexp/internal/capabilities"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"hwexp/internal/model"
)

type Adapter struct {
	smartctlPath string
}

func NewAdapter() *Adapter {
	path, _ := exec.LookPath("smartctl")
	return &Adapter{smartctlPath: path}
}

func (a *Adapter) Requirements() []capabilities.Requirement {
	return []capabilities.Requirement{
		{
			Name:        "smartctl",
			Description: "Collect SMART/NVMe health data",
			Optional:    false,
		},
	}
}

func (a *Adapter) Discover(ctx context.Context) ([]model.DiscoveredDevice, error) {
	if a.smartctlPath == "" {
		return nil, nil
	}

	devices, err := os.ReadDir("/sys/class/block")
	if err != nil {
		return nil, err
	}

	var results []model.DiscoveredDevice
	now := time.Now().UTC()

	for _, d := range devices {
		name := d.Name()
		// Only physical disks (no partitions, no loop, no ram)
		if isPhysicalDisk(name) {
			results = append(results, model.DiscoveredDevice{
				StableID:     "disk-" + name,
				Platform:     "linux",
				Source:       "linux_storage",
				DeviceClass:  "storage",
				Capabilities: []string{"smart", "health"},
				Present:      true,
				FirstSeen:    now,
				LastSeen:     now,
			})
		}
	}
	return results, nil
}

func (a *Adapter) Poll(ctx context.Context) ([]model.RawMeasurement, error) {
	if a.smartctlPath == "" {
		return nil, nil
	}

	devices, _ := a.Discover(ctx)
	var metrics []model.RawMeasurement
	now := time.Now().UTC()

	for _, d := range devices {
		name := strings.TrimPrefix(d.StableID, "disk-")
		devPath := "/dev/" + name

		cmd := exec.CommandContext(ctx, a.smartctlPath, "--json", "--all", devPath)
		out, err := cmd.Output()
		if err != nil {
			// smartctl returns non-zero exit codes for various warnings,
			// so we only error if out is empty.
			if len(out) == 0 {
				continue
			}
		}

		var smart SmartctlOutput
		if err := json.Unmarshal(out, &smart); err != nil {
			log.Printf("linux_storage: failed to parse smartctl output for %s: %v", name, err)
			continue
		}

		// 1. Health Status
		health := 0.0 // OK
		if !smart.SmartStatus.Passed {
			health = 2.0 // Critical/Failing
		}
		metrics = append(metrics, model.RawMeasurement{
			MeasurementID:  fmt.Sprintf("storage:%s:health", name),
			StableDeviceID: d.StableID,
			Source:         "linux_storage",
			RawName:        "disk_health_status",
			RawValue:       health,
			Timestamp:      now,
			Quality:        "good",
		})

		// 2. Temperature (Standardized across NVMe/SATA by smartctl)
		if smart.Temperature.Current > 0 {
			metrics = append(metrics, model.RawMeasurement{
				MeasurementID:  fmt.Sprintf("storage:%s:temp", name),
				StableDeviceID: d.StableID,
				Source:         "linux_storage",
				RawName:        "disk_temperature",
				RawValue:       float64(smart.Temperature.Current),
				Timestamp:      now,
				Quality:        "good",
			})
		}

		// 3. Life Remaining / Wear Leveling
		life := -1.0
		if smart.SmartNVMeAttributes.PercentageUsed > 0 {
			life = 100.0 - float64(smart.SmartNVMeAttributes.PercentageUsed)
		} else {
			// Search in vendor-specific attributes for SATA
			for _, attr := range smart.AtaSmartAttributes.Table {
				if attr.ID == 231 || strings.Contains(strings.ToLower(attr.Name), "wear_leveling") || strings.Contains(strings.ToLower(attr.Name), "life_left") {
					life = float64(attr.Value)
					break
				}
			}
		}
		if life >= 0 {
			metrics = append(metrics, model.RawMeasurement{
				MeasurementID:  fmt.Sprintf("storage:%s:life", name),
				StableDeviceID: d.StableID,
				Source:         "linux_storage",
				RawName:        "disk_life_remaining_percent",
				RawValue:       life,
				Timestamp:      now,
				Quality:        "good",
			})
		}

		// 4. Power On Hours
		if smart.PowerOnTime.Hours > 0 {
			metrics = append(metrics, model.RawMeasurement{
				MeasurementID:  fmt.Sprintf("storage:%s:hours", name),
				StableDeviceID: d.StableID,
				Source:         "linux_storage",
				RawName:        "disk_power_on_hours",
				RawValue:       float64(smart.PowerOnTime.Hours),
				Timestamp:      now,
				Quality:        "good",
			})
		}
	}

	return metrics, nil
}

func isPhysicalDisk(name string) bool {
	// Filter out partitions (e.g. sda1), loop, ram, and virtual devices
	if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") || strings.HasPrefix(name, "dm-") {
		return false
	}
	// Check if it's a directory in sysfs without a "partition" file
	_, err := os.Stat(filepath.Join("/sys/class/block", name, "partition"))
	return os.IsNotExist(err)
}

// Minimal Smartctl JSON structures
type SmartctlOutput struct {
	SmartStatus struct {
		Passed bool `json:"passed"`
	} `json:"smart_status"`
	Temperature struct {
		Current int `json:"current"`
	} `json:"temperature"`
	PowerOnTime struct {
		Hours int `json:"hours"`
	} `json:"power_on_time"`
	SmartNVMeAttributes struct {
		PercentageUsed int `json:"percentage_used"`
	} `json:"nvme_smart_health_information_log"`
	AtaSmartAttributes struct {
		Table []struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Value int    `json:"value"`
		} `json:"table"`
	} `json:"ata_smart_attributes"`
}
