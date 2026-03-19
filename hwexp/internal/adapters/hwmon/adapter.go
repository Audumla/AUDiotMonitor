package hwmon

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"hwexp/internal/model"
)

const DefaultBasePath = "/sys/class/hwmon"

type Adapter struct {
	basePath string
}

func NewAdapter(basePath string) *Adapter {
	if basePath == "" {
		basePath = DefaultBasePath
	}
	return &Adapter{basePath: basePath}
}

// sensorMeta maps a sensor type prefix to unit, component hint, sensor hint.
type sensorMeta struct {
	unit      string
	component string
	sensor    string
}

var sensorTypes = map[string]sensorMeta{
	"temp":     {"millidegree_celsius", "thermal", "temperature"},
	"fan":      {"rpm", "cooling", "fan_speed"},
	"power":    {"microwatt", "power", "power"},
	"in":       {"millivolt", "power", "voltage"},
	"curr":     {"milliampere", "power", "current"},
	"freq":     {"hertz", "compute", "frequency"},
	"humidity": {"millipercent", "environment", "humidity"},
	"energy":   {"microjoule", "power", "energy"},
}

var inputRE = regexp.MustCompile(`^(temp|fan|power|in|curr|freq|humidity|energy)(\d+)_input$`)

func (a *Adapter) Discover(ctx context.Context) ([]model.DiscoveredDevice, error) {
	dirs, err := filepath.Glob(filepath.Join(a.basePath, "hwmon*"))
	if err != nil {
		return nil, err
	}
	var devices []model.DiscoveredDevice
	for _, d := range dirs {
		dev, err := a.discoverOne(d)
		if err != nil {
			continue
		}
		devices = append(devices, dev)
	}
	return devices, nil
}

func (a *Adapter) discoverOne(hwmonDir string) (model.DiscoveredDevice, error) {
	driver := readSysFile(filepath.Join(hwmonDir, "name"))
	if driver == "" {
		return model.DiscoveredDevice{}, fmt.Errorf("no name file in %s", hwmonDir)
	}

	stableID, pciAddr := stableIDFor(hwmonDir, driver)
	vendor, class, subclass := classifyDriver(driver)
	// Network adapters expose hwmon with the interface name as the driver string.
	// Override the default "sensor" class when a net/ subdirectory is present.
	if class == "sensor" && isNetDevice(hwmonDir) {
		class = "network"
	}
	caps := detectCapabilities(hwmonDir)

	now := time.Now().UTC()
	return model.DiscoveredDevice{
		StableID:          stableID,
		Platform:          "linux",
		Source:            "linux_hwmon",
		DeviceClass:       class,
		DeviceSubclass:    subclass,
		Vendor:            vendor,
		Driver:            driver,
		Bus:               "pci",
		Location:          pciAddr,
		DisplayName:       driver,
		LogicalDeviceName: filepath.Base(hwmonDir),
		Capabilities:      caps,
		FirstSeen:         now,
		LastSeen:          now,
		Present:           true,
	}, nil
}

func (a *Adapter) Poll(ctx context.Context) ([]model.RawMeasurement, error) {
	dirs, err := filepath.Glob(filepath.Join(a.basePath, "hwmon*"))
	if err != nil {
		return nil, err
	}
	var all []model.RawMeasurement
	for _, d := range dirs {
		ms, _ := a.pollOne(d)
		all = append(all, ms...)
	}
	return all, nil
}

func (a *Adapter) pollOne(hwmonDir string) ([]model.RawMeasurement, error) {
	driver := readSysFile(filepath.Join(hwmonDir, "name"))
	stableID, _ := stableIDFor(hwmonDir, driver)

	entries, err := os.ReadDir(hwmonDir)
	if err != nil {
		return nil, err
	}

	var measurements []model.RawMeasurement
	now := time.Now().UTC()

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := inputRE.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}

		raw := readSysFile(filepath.Join(hwmonDir, e.Name()))
		var val float64
		if _, err := fmt.Sscanf(raw, "%f", &val); err != nil {
			continue
		}

		meta := sensorTypes[m[1]]
		measurements = append(measurements, model.RawMeasurement{
			MeasurementID:  fmt.Sprintf("linux_hwmon:%s:%s", stableID, e.Name()),
			StableDeviceID: stableID,
			Source:         "linux_hwmon",
			RawName:        e.Name(),
			RawValue:       val,
			RawUnit:        meta.unit,
			Timestamp:      now,
			Quality:        "good",
			ComponentHint:  meta.component,
			SensorHint:     meta.sensor,
		})
	}
	return measurements, nil
}

// stableIDFor derives a stable ID from PCI slot name if available, otherwise
// falls back to driver+hwmon index.
func stableIDFor(hwmonDir, driver string) (stableID, pciAddr string) {
	uevent := readSysFile(filepath.Join(hwmonDir, "device", "uevent"))
	if uevent != "" {
		scanner := bufio.NewScanner(strings.NewReader(uevent))
		for scanner.Scan() {
			if after, ok := strings.CutPrefix(scanner.Text(), "PCI_SLOT_NAME="); ok {
				pciAddr = strings.TrimSpace(after)
				return "pci-" + pciAddr, pciAddr
			}
		}
	}
	return fmt.Sprintf("hwmon-%s-%s", driver, filepath.Base(hwmonDir)), ""
}

// classifyDriver maps a hwmon driver name to vendor, device class, subclass.
func classifyDriver(driver string) (vendor, class, subclass string) {
	switch driver {
	case "amdgpu", "radeon":
		return "amd", "gpu", "discrete"
	case "nouveau", "nvidia":
		return "nvidia", "gpu", "discrete"
	case "i915", "xe":
		return "intel", "gpu", "integrated"
	case "k10temp", "k8temp", "zenpower":
		return "amd", "cpu", ""
	case "coretemp":
		return "intel", "cpu", ""
	case "acpitz":
		return "", "thermal", ""
	case "nct6775", "nct6776", "nct6779", "it87", "w83795", "w83627ehf":
		return "", "motherboard", "sensor"
	case "nvme":
		return "", "storage", "nvme"
	case "corsairpsu", "corsair-cpro", "corsaircpro":
		return "corsair", "psu", ""
	case "spd5118", "ee1004":
		return "", "memory", "spd"
	default:
		return "", "sensor", ""
	}
}

// isNetDevice returns true when the hwmon device is backed by a network interface.
// Network adapters expose a hwmon whose name is the interface name (e.g. "enp9s0",
// "iwlwifi_1") rather than a kernel driver name. The reliable indicator is the
// presence of a "net" subdirectory under the hwmon's device symlink.
func isNetDevice(hwmonDir string) bool {
	netDir := filepath.Join(hwmonDir, "device", "net")
	fi, err := os.Stat(netDir)
	return err == nil && fi.IsDir()
}

// detectCapabilities returns a deduplicated list of capability strings
// based on which sensor input files exist.
func detectCapabilities(hwmonDir string) []string {
	seen := map[string]bool{}
	entries, _ := os.ReadDir(hwmonDir)
	for _, e := range entries {
		if m := inputRE.FindStringSubmatch(e.Name()); m != nil {
			switch m[1] {
			case "temp":
				seen["temperature"] = true
			case "fan":
				seen["fan_speed"] = true
			case "power":
				seen["power"] = true
			case "in":
				seen["voltage"] = true
			case "curr":
				seen["current"] = true
			case "freq":
				seen["frequency"] = true
			}
		}
	}
	caps := make([]string, 0, len(seen))
	for k := range seen {
		caps = append(caps, k)
	}
	return caps
}

func readSysFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
