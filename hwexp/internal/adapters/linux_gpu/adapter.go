package linux_gpu

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"hwexp/internal/model"
)

const DRMBasePath = "/sys/class/drm"

type Adapter struct {
	basePath string
}

func NewAdapter(basePath string) *Adapter {
	if basePath == "" {
		basePath = DRMBasePath
	}
	return &Adapter{basePath: basePath}
}

func (a *Adapter) Discover(ctx context.Context) ([]model.DiscoveredDevice, error) {
	cards, err := filepath.Glob(filepath.Join(a.basePath, "card*"))
	if err != nil {
		return nil, err
	}

	var devices []model.DiscoveredDevice
	for _, card := range cards {
		devicePath := filepath.Join(card, "device")
		if _, err := os.Stat(devicePath); err != nil {
			continue
		}

		vendor := readSysFile(filepath.Join(devicePath, "vendor"))
		vendorName := "unknown"
		if strings.Contains(vendor, "0x1002") {
			vendorName = "amd"
		} else if strings.Contains(vendor, "0x10de") {
			vendorName = "nvidia"
		} else if strings.Contains(vendor, "0x8086") {
			vendorName = "intel"
		}

		uevent := readSysFile(filepath.Join(devicePath, "uevent"))
		stableID := "gpu-" + filepath.Base(card)
		if strings.Contains(uevent, "PCI_SLOT_NAME=") {
			for _, line := range strings.Split(uevent, "\n") {
				if strings.HasPrefix(line, "PCI_SLOT_NAME=") {
					addr := strings.TrimPrefix(line, "PCI_SLOT_NAME=")
					stableID = "pci-" + addr
					break
				}
			}
		}

		now := time.Now().UTC()
		devices = append(devices, model.DiscoveredDevice{
			StableID:          stableID,
			Platform:          "linux",
			Source:            "linux_gpu",
			DeviceClass:       "gpu",
			Vendor:            vendorName,
			DisplayName:       fmt.Sprintf("GPU %s (%s)", vendorName, filepath.Base(card)),
			LogicalDeviceName: filepath.Base(card),
			Capabilities:      []string{"utilization"},
			FirstSeen:         now,
			LastSeen:          now,
			Present:           true,
			AdapterMetadata: map[string]interface{}{
				"sysfs_path": devicePath,
			},
		})
	}

	// Also check for NVIDIA via nvidia-smi if no sysfs cards found or to supplement
	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		nvDevices, _ := discoverNvidia(ctx)
		for _, nd := range nvDevices {
			// Check if already discovered via sysfs (stableID match)
			exists := false
			for _, d := range devices {
				if d.StableID == nd.StableID {
					exists = true
					break
				}
			}
			if !exists {
				devices = append(devices, nd)
			}
		}
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

	for _, dev := range devices {
		if dev.Vendor == "amd" {
			if sysfsPath, ok := dev.AdapterMetadata["sysfs_path"].(string); ok {
				if busy := readSysFile(filepath.Join(sysfsPath, "gpu_busy_percent")); busy != "" {
					if val, err := strconv.ParseFloat(busy, 64); err == nil {
						all = append(all, model.RawMeasurement{
							MeasurementID:  fmt.Sprintf("linux_gpu:%s:gpu_busy", dev.StableID),
							StableDeviceID: dev.StableID,
							Source:         "linux_gpu",
							RawName:        "gpu_busy_percent",
							RawValue:       val,
							RawUnit:        "percent",
							Timestamp:      now,
							Quality:        "good",
							ComponentHint:  "compute",
							SensorHint:     "utilization",
						})
					}
				}
				if memBusy := readSysFile(filepath.Join(sysfsPath, "mem_busy_percent")); memBusy != "" {
					if val, err := strconv.ParseFloat(memBusy, 64); err == nil {
						all = append(all, model.RawMeasurement{
							MeasurementID:  fmt.Sprintf("linux_gpu:%s:mem_busy", dev.StableID),
							StableDeviceID: dev.StableID,
							Source:         "linux_gpu",
							RawName:        "mem_busy_percent",
							RawValue:       val,
							RawUnit:        "percent",
							Timestamp:      now,
							Quality:        "good",
							ComponentHint:  "memory",
							SensorHint:     "utilization",
						})
					}
				}
			}
		} else if dev.Vendor == "nvidia" {
			ms, _ := pollNvidia(ctx, dev.StableID)
			all = append(all, ms...)
		}
	}

	return all, nil
}

func discoverNvidia(ctx context.Context) ([]model.DiscoveredDevice, error) {
	out, err := exec.CommandContext(ctx, "nvidia-smi", "--query-gpu=pci.bus_id,name", "--format=csv,noheader,nounits").Output()
	if err != nil {
		return nil, err
	}

	var devices []model.DiscoveredDevice
	now := time.Now().UTC()
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ",")
		if len(parts) < 2 {
			continue
		}
		pciID := strings.TrimSpace(parts[0])
		name := strings.TrimSpace(parts[1])
		stableID := "pci-" + strings.ToLower(pciID)

		devices = append(devices, model.DiscoveredDevice{
			StableID:     stableID,
			Platform:     "linux",
			Source:       "nvidia_smi",
			DeviceClass:  "gpu",
			Vendor:       "nvidia",
			Model:        name,
			DisplayName:  name,
			Capabilities: []string{"utilization"},
			FirstSeen:    now,
			LastSeen:     now,
			Present:      true,
		})
	}
	return devices, nil
}

func pollNvidia(ctx context.Context, stableID string) ([]model.RawMeasurement, error) {
	// StableID is pci-0000:0b:00.0, nvidia-smi wants the 0000:0b:00.0 part
	pciAddr := strings.TrimPrefix(stableID, "pci-")
	out, err := exec.CommandContext(ctx, "nvidia-smi", "--id="+pciAddr, "--query-gpu=utilization.gpu,utilization.memory", "--format=csv,noheader,nounits").Output()
	if err != nil {
		return nil, err
	}

	var ms []model.RawMeasurement
	now := time.Now().UTC()
	parts := strings.Split(strings.TrimSpace(string(out)), ",")
	if len(parts) >= 2 {
		if gpuBusy, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64); err == nil {
			ms = append(ms, model.RawMeasurement{
				MeasurementID:  fmt.Sprintf("nvidia_smi:%s:gpu_busy", stableID),
				StableDeviceID: stableID,
				Source:         "nvidia_smi",
				RawName:        "gpu_busy_percent",
				RawValue:       gpuBusy,
				RawUnit:        "percent",
				Timestamp:      now,
				Quality:        "good",
				ComponentHint:  "compute",
				SensorHint:     "utilization",
			})
		}
		if memBusy, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
			ms = append(ms, model.RawMeasurement{
				MeasurementID:  fmt.Sprintf("nvidia_smi:%s:mem_busy", stableID),
				StableDeviceID: stableID,
				Source:         "nvidia_smi",
				RawName:        "mem_busy_percent",
				RawValue:       memBusy,
				RawUnit:        "percent",
				Timestamp:      now,
				Quality:        "good",
				ComponentHint:  "memory",
				SensorHint:     "utilization",
			})
		}
	}
	return ms, nil
}

func readSysFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
