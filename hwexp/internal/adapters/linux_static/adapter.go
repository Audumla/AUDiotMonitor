package linux_static

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"hwexp/internal/model"
)

const DMIBasePath = "/sys/class/dmi/id"

type Adapter struct {
	basePath string
}

func NewAdapter(basePath string) *Adapter {
	if basePath == "" {
		basePath = DMIBasePath
	}
	return &Adapter{basePath: basePath}
}

func (a *Adapter) Discover(ctx context.Context) ([]model.DiscoveredDevice, error) {
	var devices []model.DiscoveredDevice
	now := time.Now().UTC()

	// 1. Motherboard / System
	boardVendor := readSysFile(filepath.Join(a.basePath, "board_vendor"))
	boardName := readSysFile(filepath.Join(a.basePath, "board_name"))
	if boardName != "" {
		devices = append(devices, model.DiscoveredDevice{
			StableID:     "system-motherboard",
			Platform:     "linux",
			Source:       "linux_static",
			DeviceClass:  "motherboard",
			Vendor:       boardVendor,
			Model:        boardName,
			DisplayName:  fmt.Sprintf("Motherboard: %s %s", boardVendor, boardName),
			Capabilities: []string{"inventory"},
			Present:      true,
			FirstSeen:    now,
			LastSeen:     now,
			AdapterMetadata: map[string]interface{}{
				"bios_version": readSysFile(filepath.Join(a.basePath, "bios_version")),
				"bios_date":    readSysFile(filepath.Join(a.basePath, "bios_date")),
				"sys_vendor":   readSysFile(filepath.Join(a.basePath, "sys_vendor")),
				"product_name": readSysFile(filepath.Join(a.basePath, "product_name")),
			},
		})
	}

	// 2. CPU Specs
	cpuModel, cores, threads := getCPUInfo()
	if cpuModel != "" {
		devices = append(devices, model.DiscoveredDevice{
			StableID:     "system-cpu",
			Platform:     "linux",
			Source:       "linux_static",
			DeviceClass:  "cpu",
			Model:        cpuModel,
			DisplayName:  cpuModel,
			Capabilities: []string{"inventory"},
			Present:      true,
			FirstSeen:    now,
			LastSeen:     now,
			AdapterMetadata: map[string]interface{}{
				"cores":   cores,
				"threads": threads,
			},
		})
	}

	// 3. Memory Specs
	memTotal := getTotalRAM()
	if memTotal > 0 {
		devices = append(devices, model.DiscoveredDevice{
			StableID:     "system-mem",
			Platform:     "linux",
			Source:       "linux_static",
			DeviceClass:  "memory",
			DisplayName:  "System Memory",
			Capabilities: []string{"inventory"},
			Present:      true,
			FirstSeen:    now,
			LastSeen:     now,
			AdapterMetadata: map[string]interface{}{
				"capacity_bytes": memTotal,
			},
		})
	}

	return devices, nil
}

func (a *Adapter) Poll(ctx context.Context) ([]model.RawMeasurement, error) {
	devices, _ := a.Discover(ctx)
	var all []model.RawMeasurement
	now := time.Now().UTC()

	for _, dev := range devices {
		if dev.DeviceClass == "cpu" {
			if c, ok := dev.AdapterMetadata["cores"].(int); ok {
				all = append(all, model.RawMeasurement{
					MeasurementID:  "linux_static:cpu:cores",
					StableDeviceID: dev.StableID,
					Source:         "linux_static",
					RawName:        "cpu_cores_count",
					RawValue:       float64(c),
					RawUnit:        "count",
					Timestamp:      now,
					Quality:        "good",
					ComponentHint:  "compute",
					SensorHint:     "cores",
				})
			}
			if t, ok := dev.AdapterMetadata["threads"].(int); ok {
				all = append(all, model.RawMeasurement{
					MeasurementID:  "linux_static:cpu:threads",
					StableDeviceID: dev.StableID,
					Source:         "linux_static",
					RawName:        "cpu_threads_count",
					RawValue:       float64(t),
					RawUnit:        "count",
					Timestamp:      now,
					Quality:        "good",
					ComponentHint:  "compute",
					SensorHint:     "threads",
				})
			}
		} else if dev.DeviceClass == "memory" {
			if cap, ok := dev.AdapterMetadata["capacity_bytes"].(float64); ok {
				all = append(all, model.RawMeasurement{
					MeasurementID:  "linux_static:mem:capacity",
					StableDeviceID: dev.StableID,
					Source:         "linux_static",
					RawName:        "memory_capacity_bytes",
					RawValue:       cap,
					RawUnit:        "bytes",
					Timestamp:      now,
					Quality:        "good",
					ComponentHint:  "memory",
					SensorHint:     "capacity",
				})
			}
		}
	}

	return all, nil
}

func getCPUInfo() (modelName string, cores, threads int) {
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return "", 0, 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	coreMap := make(map[string]bool)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "model name") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				modelName = strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "processor") {
			threads++
		}
		if strings.HasPrefix(line, "core id") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				coreMap[strings.TrimSpace(parts[1])] = true
			}
		}
	}
	cores = len(coreMap)
	if cores == 0 { cores = threads } // Fallback for some virtualized environments
	return
}

func getTotalRAM() float64 {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		line := scanner.Text() // first line is MemTotal
		re := regexp.MustCompile(`MemTotal:\s+(\d+) kB`)
		m := re.FindStringSubmatch(line)
		if len(m) > 1 {
			kb, _ := strconv.ParseFloat(m[1], 64)
			return kb * 1024
		}
	}
	return 0
}

func readSysFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
