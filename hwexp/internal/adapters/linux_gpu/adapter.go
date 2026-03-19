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
	"hwexp/internal/pcidb"
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

		vendorHex := readSysFile(filepath.Join(devicePath, "vendor"))
		deviceHex := readSysFile(filepath.Join(devicePath, "device"))
		vendorName := "unknown"
		if strings.Contains(vendorHex, "1002") {
			vendorName = "amd"
		} else if strings.Contains(vendorHex, "10de") {
			vendorName = "nvidia"
		} else if strings.Contains(vendorHex, "8086") {
			vendorName = "intel"
		}

		// Build raw identifier
		pciVendorHex := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(vendorHex)), "0x")
		pciDeviceHex := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(deviceHex)), "0x")
		rawIDs := map[string]string{}
		if pciVendorHex != "" && pciDeviceHex != "" {
			rawIDs["pci_id"] = pciVendorHex + ":" + pciDeviceHex
		}

		uevent := readSysFile(filepath.Join(devicePath, "uevent"))
		stableID := "gpu-" + filepath.Base(card)
		if strings.Contains(uevent, "PCI_SLOT_NAME=") {
			for _, line := range strings.Split(uevent, "\n") {
				if strings.HasPrefix(line, "PCI_SLOT_NAME=") {
					addr := strings.TrimPrefix(line, "PCI_SLOT_NAME=")
					stableID = "pci-" + addr
					rawIDs["pci_slot"] = addr
					break
				}
			}
		}

		now := time.Now().UTC()
		gpuModel := ""
		meta := map[string]interface{}{
			"sysfs_path": devicePath,
		}

		if vendorName == "amd" {
			if vramRaw := readSysFile(filepath.Join(devicePath, "mem_info_vram_total")); vramRaw != "" {
				if v, err := strconv.ParseFloat(vramRaw, 64); err == nil {
					meta["capacity_bytes"] = v
				}
			}
			if gttSize := readSysFile(filepath.Join(devicePath, "mem_info_gtt_size")); gttSize != "" {
				if v, err := strconv.ParseFloat(gttSize, 64); err == nil && v > 0 {
					meta["gtt_size_bytes"] = v
				}
			}
			if vramVendor := readSysFile(filepath.Join(devicePath, "mem_info_vram_vendor")); vramVendor != "" {
				meta["vram_vendor"] = vramVendor
			}
			gpuModel = readSysFile(filepath.Join(devicePath, "label"))
			if gpuModel == "" {
				gpuModel = readSysFile(filepath.Join(devicePath, "product_name"))
			}
		}

		// Fall back to pci.ids lookup for model name
		if gpuModel == "" && pciVendorHex != "" && pciDeviceHex != "" {
			_, gpuModel = pcidb.Lookup(pciVendorHex, pciDeviceHex)
		}

		displayName := fmt.Sprintf("GPU %s (%s)", vendorName, filepath.Base(card))
		if gpuModel != "" {
			displayName = gpuModel
		}

		devices = append(devices, model.DiscoveredDevice{
			StableID:          stableID,
			Platform:          "linux",
			Source:            "linux_gpu",
			DeviceClass:       "gpu",
			Vendor:            vendorName,
			Model:             gpuModel,
			DisplayName:       displayName,
			LogicalDeviceName: filepath.Base(card),
			Capabilities:      []string{"utilization", "inventory"},
			RawIdentifiers:    rawIDs,
			FirstSeen:         now,
			LastSeen:          now,
			Present:           true,
			AdapterMetadata:   meta,
		})
	}

	// Also check for NVIDIA via nvidia-smi if no sysfs cards found or to supplement
	if _, err := exec.LookPath("nvidia-smi"); err == nil {
		nvDevices, _ := discoverNvidia(ctx)
		for _, nd := range nvDevices {
			// Check if already discovered via sysfs (stableID match)
			exists := false
			for i, d := range devices {
				if d.StableID == nd.StableID {
					// Supplement sysfs info with nvidia-smi info if available
					if d.Vendor == "nvidia" {
						devices[i].Capabilities = append(devices[i].Capabilities, "inventory")
						devices[i].AdapterMetadata["capacity_bytes"] = nd.AdapterMetadata["capacity_bytes"]
						devices[i].Model = nd.Model
						devices[i].DisplayName = nd.DisplayName
					}
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
		if cap, ok := dev.AdapterMetadata["capacity_bytes"].(float64); ok && cap > 0 {
			all = append(all, model.RawMeasurement{
				MeasurementID:  fmt.Sprintf("linux_gpu:%s:vram_capacity", dev.StableID),
				StableDeviceID: dev.StableID,
				Source:         "linux_gpu",
				RawName:        "vram_capacity_bytes",
				RawValue:       cap,
				RawUnit:        "bytes",
				Timestamp:      now,
				Quality:        "good",
				ComponentHint:  "memory",
				SensorHint:     "capacity",
			})
		}

		if dev.Vendor == "amd" {
			if sysfsPath, ok := dev.AdapterMetadata["sysfs_path"].(string); ok {
				// Prefer the binary gpu_metrics file (atomic SMU snapshot, supported
				// on all RDNA/RDNA2/RDNA3 cards). Fall back to the per-file sysfs
				// text interface when gpu_metrics is absent (older kernels/drivers).
				gm := readGPUMetrics(sysfsPath)
				if gm.Valid {
					all = append(all,
						model.RawMeasurement{
							MeasurementID:  fmt.Sprintf("linux_gpu:%s:gpu_busy_percent", dev.StableID),
							StableDeviceID: dev.StableID,
							Source:         "linux_gpu",
							RawName:        "gpu_busy_percent",
							RawValue:       float64(gm.GFXActivity),
							RawUnit:        "percent",
							Timestamp:      now,
							Quality:        "good",
							ComponentHint:  "compute",
							SensorHint:     "utilization",
						},
						model.RawMeasurement{
							MeasurementID:  fmt.Sprintf("linux_gpu:%s:mem_busy_percent", dev.StableID),
							StableDeviceID: dev.StableID,
							Source:         "linux_gpu",
							RawName:        "mem_busy_percent",
							RawValue:       float64(gm.UMCActivity),
							RawUnit:        "percent",
							Timestamp:      now,
							Quality:        "good",
							ComponentHint:  "memory",
							SensorHint:     "utilization",
						},
					)
					if gm.MMActivity > 0 {
						all = append(all, model.RawMeasurement{
							MeasurementID:  fmt.Sprintf("linux_gpu:%s:mm_busy_percent", dev.StableID),
							StableDeviceID: dev.StableID,
							Source:         "linux_gpu",
							RawName:        "mm_busy_percent",
							RawValue:       float64(gm.MMActivity),
							RawUnit:        "percent",
							Timestamp:      now,
							Quality:        "good",
							ComponentHint:  "media",
							SensorHint:     "utilization",
						})
					}
				} else {
					// Fallback: plain sysfs text files
					for _, m := range []struct {
						file, name, unit, comp, sensor string
					}{
						{"gpu_busy_percent", "gpu_busy_percent", "percent", "compute", "utilization"},
						{"mem_busy_percent", "mem_busy_percent", "percent", "memory", "utilization"},
					} {
						if raw := readSysFile(filepath.Join(sysfsPath, m.file)); raw != "" {
							if val, err := strconv.ParseFloat(raw, 64); err == nil {
								all = append(all, model.RawMeasurement{
									MeasurementID:  fmt.Sprintf("linux_gpu:%s:%s", dev.StableID, m.file),
									StableDeviceID: dev.StableID,
									Source:         "linux_gpu",
									RawName:        m.name,
									RawValue:       val,
									RawUnit:        m.unit,
									Timestamp:      now,
									Quality:        "good",
									ComponentHint:  m.comp,
									SensorHint:     m.sensor,
								})
							}
						}
					}
				}

				// VRAM / GTT used — always from sysfs text files (not in gpu_metrics)
				vramUsed := 0.0
				vramTotal := 0.0
				for _, m := range []struct {
					file, name, unit, comp, sensor string
				}{
					{"mem_info_vram_used", "vram_used_bytes", "bytes", "memory", "used"},
					{"mem_info_gtt_used", "gtt_used_bytes", "bytes", "memory", "used"},
					{"mem_info_vram_total", "vram_total_bytes", "bytes", "memory", "capacity"},
				} {
					if raw := readSysFile(filepath.Join(sysfsPath, m.file)); raw != "" {
						if val, err := strconv.ParseFloat(raw, 64); err == nil {
							if m.file == "mem_info_vram_used" { vramUsed = val }
							if m.file == "mem_info_vram_total" { vramTotal = val }
							all = append(all, model.RawMeasurement{
								MeasurementID:  fmt.Sprintf("linux_gpu:%s:%s", dev.StableID, m.file),
								StableDeviceID: dev.StableID,
								Source:         "linux_gpu",
								RawName:        m.name,
								RawValue:       val,
								RawUnit:        m.unit,
								Timestamp:      now,
								Quality:        "good",
								ComponentHint:  m.comp,
								SensorHint:     m.sensor,
							})
						}
					}
				}

				if vramTotal > 0 {
					all = append(all, model.RawMeasurement{
						MeasurementID:  fmt.Sprintf("linux_gpu:%s:vram_usage_percent", dev.StableID),
						StableDeviceID: dev.StableID,
						Source:         "linux_gpu",
						RawName:        "vram_usage_percent",
						RawValue:       (vramUsed / vramTotal) * 100.0,
						RawUnit:        "percent",
						Timestamp:      now,
						Quality:        "good",
						ComponentHint:  "memory",
						SensorHint:     "usage",
					})
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
	out, err := exec.CommandContext(ctx, "nvidia-smi", "--id="+pciAddr, "--query-gpu=utilization.gpu,utilization.memory,memory.used,memory.total", "--format=csv,noheader,nounits").Output()
	if err != nil {
		return nil, err
	}

	var ms []model.RawMeasurement
	now := time.Now().UTC()
	parts := strings.Split(strings.TrimSpace(string(out)), ",")
	if len(parts) >= 4 {
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
		memUsed, err1 := strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
		memTotal, err2 := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		if err1 == nil && err2 == nil && memTotal > 0 {
			ms = append(ms, model.RawMeasurement{
				MeasurementID:  fmt.Sprintf("nvidia_smi:%s:vram_usage", stableID),
				StableDeviceID: stableID,
				Source:         "nvidia_smi",
				RawName:        "vram_usage_percent",
				RawValue:       (memUsed / memTotal) * 100.0,
				RawUnit:        "percent",
				Timestamp:      now,
				Quality:        "good",
				ComponentHint:  "memory",
				SensorHint:     "usage",
			})
			ms = append(ms, model.RawMeasurement{
				MeasurementID:  fmt.Sprintf("nvidia_smi:%s:vram_used_bytes", stableID),
				StableDeviceID: stableID,
				Source:         "nvidia_smi",
				RawName:        "vram_used_bytes",
				RawValue:       memUsed * 1024 * 1024, // nvidia-smi returns MiB
				RawUnit:        "bytes",
				Timestamp:      now,
				Quality:        "good",
				ComponentHint:  "memory",
				SensorHint:     "used",
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
