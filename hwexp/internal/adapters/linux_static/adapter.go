package linux_static

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

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
		meta := map[string]interface{}{
			"bios_version": readSysFile(filepath.Join(a.basePath, "bios_version")),
			"bios_date":    readSysFile(filepath.Join(a.basePath, "bios_date")),
			"sys_vendor":   readSysFile(filepath.Join(a.basePath, "sys_vendor")),
			"product_name": readSysFile(filepath.Join(a.basePath, "product_name")),
		}
		// Only include DMI fields that have real values — filter OEM placeholders
		for k, path := range map[string]string{
			"product_serial":  filepath.Join(a.basePath, "product_serial"),
			"product_uuid":    filepath.Join(a.basePath, "product_uuid"),
			"chassis_type":    filepath.Join(a.basePath, "chassis_type"),
			"chassis_vendor":  filepath.Join(a.basePath, "chassis_vendor"),
			"chassis_serial":  filepath.Join(a.basePath, "chassis_serial"),
			"chassis_version": filepath.Join(a.basePath, "chassis_version"),
		} {
			v := readDMIFile(path)
			if k == "chassis_type" {
				v = chassisTypeStr(v)
			}
			if v != "" {
				meta[k] = v
			}
		}
		devices = append(devices, model.DiscoveredDevice{
			StableID:        "system-motherboard",
			Platform:        "linux",
			Source:          "linux_static",
			DeviceClass:     "motherboard",
			Vendor:          boardVendor,
			Model:           boardName,
			DisplayName:     fmt.Sprintf("Motherboard: %s %s", boardVendor, boardName),
			Capabilities:    []string{"inventory"},
			Present:         true,
			FirstSeen:       now,
			LastSeen:        now,
			AdapterMetadata: meta,
		})
	}

	// 2. CPU
	cpuModel, cores, threads := getCPUInfo()
	if cpuModel != "" {
		meta := map[string]interface{}{
			"cores":   cores,
			"threads": threads,
		}
		if maxFreq := readSysFile("/sys/devices/system/cpu/cpu0/cpufreq/cpuinfo_max_freq"); maxFreq != "" {
			if khz, err := strconv.ParseFloat(maxFreq, 64); err == nil {
				meta["max_freq_hz"] = khz * 1000
			}
		}
		if governor := readSysFile("/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor"); governor != "" {
			meta["governor"] = governor
		}
		// Turbo/boost — Intel pstate: no_turbo=0 means boost ON; generic: boost=1 means ON
		if noTurbo := readSysFile("/sys/devices/system/cpu/intel_pstate/no_turbo"); noTurbo != "" {
			meta["turbo_enabled"] = noTurbo == "0"
		} else if boost := readSysFile("/sys/devices/system/cpu/cpufreq/boost"); boost != "" {
			meta["turbo_enabled"] = boost == "1"
		}
		for k, v := range getCPUCacheInfo() {
			meta[k] = v
		}
		devices = append(devices, model.DiscoveredDevice{
			StableID:        "system-cpu",
			Platform:        "linux",
			Source:          "linux_static",
			DeviceClass:     "cpu",
			Model:           cpuModel,
			DisplayName:     cpuModel,
			Capabilities:    []string{"inventory", "frequency"},
			Present:         true,
			FirstSeen:       now,
			LastSeen:        now,
			AdapterMetadata: meta,
		})
	}

	// 3. Memory
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

	// 3b. Memory DIMMs (via dmidecode, cached after first run)
	devices = append(devices, getDIMMDevices(now)...)

	// 4. NVMe drives
	devices = append(devices, getNVMeDevices(now)...)

	// 5. Network interfaces
	devices = append(devices, getNetDevices(now)...)

	// 6. SATA / SAS block devices
	devices = append(devices, getBlockDevices(now)...)

	return devices, nil
}

func (a *Adapter) Poll(ctx context.Context) ([]model.RawMeasurement, error) {
	devices, _ := a.Discover(ctx)
	var all []model.RawMeasurement
	now := time.Now().UTC()

	for _, dev := range devices {
		switch dev.DeviceClass {

		case "cpu":
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
			// Average current frequency across all online CPUs
			if avgHz := getCPUCurFreqHz(); avgHz > 0 {
				all = append(all, model.RawMeasurement{
					MeasurementID:  "linux_static:cpu:cur_freq",
					StableDeviceID: dev.StableID,
					Source:         "linux_static",
					RawName:        "cpu_freq_hz",
					RawValue:       avgHz,
					RawUnit:        "hertz",
					Timestamp:      now,
					Quality:        "good",
					ComponentHint:  "compute",
					SensorHint:     "frequency",
				})
			}

		case "memory":
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

		case "network":
			iface := dev.LogicalDeviceName
			if iface == "" {
				continue
			}
			statsDir := filepath.Join("/sys/class/net", iface, "statistics")
			for _, s := range []struct {
				file, name, unit, comp, sensor string
			}{
				{"rx_bytes", "net_rx_bytes_total", "bytes", "network", "rx_bytes"},
				{"tx_bytes", "net_tx_bytes_total", "bytes", "network", "tx_bytes"},
				{"rx_packets", "net_rx_packets_total", "count", "network", "rx_packets"},
				{"tx_packets", "net_tx_packets_total", "count", "network", "tx_packets"},
				{"rx_errors", "net_rx_errors_total", "count", "network", "rx_errors"},
				{"tx_errors", "net_tx_errors_total", "count", "network", "tx_errors"},
				{"rx_dropped", "net_rx_dropped_total", "count", "network", "rx_dropped"},
				{"tx_dropped", "net_tx_dropped_total", "count", "network", "tx_dropped"},
			} {
				raw := readSysFile(filepath.Join(statsDir, s.file))
				if raw == "" {
					continue
				}
				val, err := strconv.ParseFloat(raw, 64)
				if err != nil {
					continue
				}
				all = append(all, model.RawMeasurement{
					MeasurementID:  fmt.Sprintf("linux_static:net:%s:%s", iface, s.file),
					StableDeviceID: dev.StableID,
					Source:         "linux_static",
					RawName:        s.name,
					RawValue:       val,
					RawUnit:        s.unit,
					Timestamp:      now,
					Quality:        "good",
					ComponentHint:  s.comp,
					SensorHint:     s.sensor,
				})
			}
		}
	}

	return all, nil
}

// ── Discovery helpers ────────────────────────────────────────────────────────

func getNVMeDevices(now time.Time) []model.DiscoveredDevice {
	dirs, _ := filepath.Glob("/sys/class/nvme/nvme*")
	var devices []model.DiscoveredDevice
	for i, d := range dirs {
		modelStr := strings.TrimSpace(readSysFile(filepath.Join(d, "model")))
		if modelStr == "" {
			continue
		}
		serial := strings.TrimSpace(readSysFile(filepath.Join(d, "serial")))
		firmware := strings.TrimSpace(readSysFile(filepath.Join(d, "firmware_rev")))
		addr := strings.TrimSpace(readSysFile(filepath.Join(d, "address")))

		stableID := fmt.Sprintf("system-nvme-%d", i)
		if addr != "" {
			stableID = "pci-" + addr
		}

		vendor := nvmeVendorFromModel(modelStr)

		devices = append(devices, model.DiscoveredDevice{
			StableID:       stableID,
			Platform:       "linux",
			Source:         "linux_static",
			DeviceClass:    "storage",
			DeviceSubclass: "nvme",
			Vendor:         vendor,
			Model:          modelStr,
			DisplayName:    modelStr,
			Capabilities:   []string{"inventory"},
			Present:        true,
			FirstSeen:      now,
			LastSeen:       now,
			AdapterMetadata: map[string]interface{}{
				"serial":       serial,
				"firmware_rev": firmware,
			},
		})
	}
	return devices
}

func getNetDevices(now time.Time) []model.DiscoveredDevice {
	dirs, _ := filepath.Glob("/sys/class/net/*")
	var devices []model.DiscoveredDevice
	for _, d := range dirs {
		iface := filepath.Base(d)
		// Skip loopback and virtual interfaces (no "device" symlink)
		if iface == "lo" {
			continue
		}
		deviceLink := filepath.Join(d, "device")
		if _, err := os.Stat(deviceLink); err != nil {
			continue
		}

		mac := readSysFile(filepath.Join(d, "address"))
		operstate := readSysFile(filepath.Join(d, "operstate"))
		mtu := readSysFile(filepath.Join(d, "mtu"))

		// Derive a stable ID from the underlying PCI address when possible.
		// This causes the network device to merge with the hwmon entry for
		// the same NIC (which also uses pci-<addr> as its stable ID).
		stableID := "net-" + iface
		if realPath, err := filepath.EvalSymlinks(deviceLink); err == nil {
			if base := filepath.Base(realPath); isPCIAddress(base) {
				stableID = "pci-" + base
			}
		}

		meta := map[string]interface{}{
			"operstate": operstate,
		}
		if mac != "" {
			meta["mac"] = mac
		}
		if mtu != "" {
			meta["mtu"] = mtu
		}
		if speedStr := readSysFile(filepath.Join(d, "speed")); speedStr != "" {
			if speedMbps, err := strconv.ParseInt(speedStr, 10, 64); err == nil && speedMbps > 0 {
				meta["speed_mbps"] = speedMbps
			}
		}
		if duplex := readSysFile(filepath.Join(d, "duplex")); duplex != "" {
			meta["duplex"] = duplex
		}

		devices = append(devices, model.DiscoveredDevice{
			StableID:          stableID,
			Platform:          "linux",
			Source:            "linux_static",
			DeviceClass:       "network",
			LogicalDeviceName: iface,
			DisplayName:       iface,
			Capabilities:      []string{"inventory", "statistics"},
			Present:           true,
			FirstSeen:         now,
			LastSeen:          now,
			AdapterMetadata:   meta,
		})
	}
	return devices
}

func getBlockDevices(now time.Time) []model.DiscoveredDevice {
	dirs, _ := filepath.Glob("/sys/class/block/sd*")
	var devices []model.DiscoveredDevice
	for _, d := range dirs {
		// Skip partitions — whole disks have no "partition" file
		if _, err := os.Stat(filepath.Join(d, "partition")); err == nil {
			continue
		}
		deviceDir := filepath.Join(d, "device")
		if _, err := os.Stat(deviceDir); err != nil {
			continue
		}

		vendor := strings.TrimSpace(readSysFile(filepath.Join(deviceDir, "vendor")))
		modelStr := strings.TrimSpace(readSysFile(filepath.Join(deviceDir, "model")))
		rev := strings.TrimSpace(readSysFile(filepath.Join(deviceDir, "rev")))

		sizeBytes := 0.0
		if sizeStr := readSysFile(filepath.Join(d, "size")); sizeStr != "" {
			if sectors, err := strconv.ParseFloat(sizeStr, 64); err == nil {
				sizeBytes = sectors * 512
			}
		}

		subclass := "ssd"
		if readSysFile(filepath.Join(d, "queue", "rotational")) == "1" {
			subclass = "hdd"
		}

		displayName := strings.TrimSpace(vendor + " " + modelStr)
		if displayName == "" {
			displayName = filepath.Base(d)
		}

		// Derive stable ID from SCSI H:C:T:L address embedded in the sysfs path
		stableID := "block-" + filepath.Base(d)
		if realPath, err := filepath.EvalSymlinks(d); err == nil {
			parts := strings.Split(realPath, "/")
			for i, p := range parts {
				if strings.HasPrefix(p, "target") && i+1 < len(parts) {
					stableID = "scsi-" + parts[i+1]
					break
				}
			}
		}

		meta := map[string]interface{}{
			"size_bytes":  sizeBytes,
			"rotational":  subclass == "hdd",
		}
		if rev != "" {
			meta["firmware_rev"] = rev
		}

		devices = append(devices, model.DiscoveredDevice{
			StableID:       stableID,
			Platform:       "linux",
			Source:         "linux_static",
			DeviceClass:    "storage",
			DeviceSubclass: subclass,
			Vendor:         blockVendorNorm(vendor),
			Model:          modelStr,
			DisplayName:    displayName,
			Capabilities:   []string{"inventory"},
			Present:        true,
			FirstSeen:      now,
			LastSeen:       now,
			AdapterMetadata: meta,
		})
	}
	return devices
}

// ── Poll helpers ─────────────────────────────────────────────────────────────

// getCPUCurFreqHz returns the average current scaling frequency across all
// online CPUs, in Hz. Returns 0 if cpufreq is unavailable (e.g. in VMs).
func getCPUCurFreqHz() float64 {
	files, _ := filepath.Glob("/sys/devices/system/cpu/cpu*/cpufreq/scaling_cur_freq")
	var total float64
	count := 0
	for _, f := range files {
		if raw := readSysFile(f); raw != "" {
			if khz, err := strconv.ParseFloat(raw, 64); err == nil {
				total += khz * 1000 // kHz → Hz
				count++
			}
		}
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

// ── Static info helpers ───────────────────────────────────────────────────────

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
	if cores == 0 {
		cores = threads // fallback for virtualised environments
	}
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
		line := scanner.Text()
		re := regexp.MustCompile(`MemTotal:\s+(\d+) kB`)
		m := re.FindStringSubmatch(line)
		if len(m) > 1 {
			kb, _ := strconv.ParseFloat(m[1], 64)
			return kb * 1024
		}
	}
	return 0
}

// isPCIAddress returns true for strings matching the PCI slot format 0000:09:00.0
func isPCIAddress(s string) bool {
	return len(s) == 12 && s[4] == ':' && s[7] == ':' && s[10] == '.'
}

// nvmeVendorFromModel infers vendor from the NVMe model string.
func nvmeVendorFromModel(modelStr string) string {
	ml := strings.ToLower(modelStr)
	switch {
	case strings.Contains(ml, "samsung"):
		return "samsung"
	case strings.Contains(ml, "western digital") || strings.HasPrefix(ml, "wd"):
		return "western_digital"
	case strings.Contains(ml, "seagate"):
		return "seagate"
	case strings.Contains(ml, "kingston"):
		return "kingston"
	case strings.Contains(ml, "crucial"):
		return "crucial"
	case strings.Contains(ml, "sk hynix") || strings.Contains(ml, "skhynix"):
		return "sk_hynix"
	case strings.Contains(ml, "micron"):
		return "micron"
	case strings.Contains(ml, "intel"):
		return "intel"
	case strings.Contains(ml, "sandisk"):
		return "sandisk"
	case strings.Contains(ml, "viper") || strings.Contains(ml, "patriot"):
		return "patriot"
	case strings.Contains(ml, "sabrent"):
		return "sabrent"
	case strings.Contains(ml, "teamgroup") || strings.Contains(ml, "team group"):
		return "teamgroup"
	}
	return ""
}

// blockVendorNorm normalises the vendor string read from SCSI/ATA device attributes.
func blockVendorNorm(vendor string) string {
	v := strings.ToLower(strings.TrimSpace(vendor))
	switch {
	case strings.Contains(v, "samsung"):
		return "samsung"
	case strings.Contains(v, "seagate") || v == "st":
		return "seagate"
	case strings.Contains(v, "western") || v == "wdc" || v == "wd":
		return "western_digital"
	case strings.Contains(v, "toshiba"):
		return "toshiba"
	case strings.Contains(v, "hitachi") || v == "hgst":
		return "hitachi"
	case strings.Contains(v, "intel"):
		return "intel"
	case strings.Contains(v, "crucial") || strings.Contains(v, "micron"):
		return "micron"
	case v == "ata":
		return "" // generic ATA placeholder — not useful
	}
	return vendor
}

var chassisTypes = map[string]string{
	"1": "Other", "2": "Unknown", "3": "Desktop", "4": "Low Profile Desktop",
	"5": "Pizza Box", "6": "Mini Tower", "7": "Tower", "8": "Portable",
	"9": "Laptop", "10": "Notebook", "11": "Hand Held", "12": "Docking Station",
	"13": "All In One", "14": "Sub Notebook", "15": "Space-saving",
	"17": "Main Server Chassis", "22": "RAID Chassis", "23": "Rack Mount Chassis",
	"28": "Blade", "30": "Tablet", "31": "Convertible", "32": "Detachable",
	"35": "Mini PC", "36": "Stick PC",
}

// chassisTypeStr maps the raw SMBIOS chassis type byte to a human-readable string.
func chassisTypeStr(raw string) string {
	if raw == "" {
		return ""
	}
	n := strings.TrimSpace(raw)
	if strings.HasPrefix(n, "0x") || strings.HasPrefix(n, "0X") {
		if v, err := strconv.ParseInt(n[2:], 16, 32); err == nil {
			n = strconv.FormatInt(v, 10)
		}
	}
	if label, ok := chassisTypes[n]; ok {
		return label
	}
	return raw
}

func readSysFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// getCPUCacheInfo reads the CPU cache topology from sysfs and returns a map
// of cache level keys to size strings (e.g. "l1_cache_data" -> "48K").
func getCPUCacheInfo() map[string]interface{} {
	result := map[string]interface{}{}
	dirs, _ := filepath.Glob("/sys/devices/system/cpu/cpu0/cache/index*")
	for _, d := range dirs {
		level := readSysFile(filepath.Join(d, "level"))
		cacheType := readSysFile(filepath.Join(d, "type"))
		size := readSysFile(filepath.Join(d, "size"))
		if level == "" || cacheType == "" || size == "" {
			continue
		}
		var key string
		switch cacheType {
		case "Unified":
			key = fmt.Sprintf("l%s_cache", level)
		case "Instruction":
			key = fmt.Sprintf("l%s_cache_instruction", level)
		case "Data":
			key = fmt.Sprintf("l%s_cache_data", level)
		default:
			key = fmt.Sprintf("l%s_cache_%s", level, strings.ToLower(cacheType))
		}
		result[key] = size
	}
	return result
}

// DIMM discovery via dmidecode — cached after first successful run.
var (
	dimmOnce    sync.Once
	dimmDevices []model.DiscoveredDevice
)

func getDIMMDevices(now time.Time) []model.DiscoveredDevice {
	dimmOnce.Do(func() {
		dimmDevices = runDMIDecode()
	})
	if len(dimmDevices) == 0 {
		return nil
	}
	// Return copies with updated LastSeen
	out := make([]model.DiscoveredDevice, len(dimmDevices))
	copy(out, dimmDevices)
	for i := range out {
		out[i].LastSeen = now
	}
	return out
}

func runDMIDecode() []model.DiscoveredDevice {
	if _, err := exec.LookPath("dmidecode"); err != nil {
		return nil
	}
	out, err := exec.Command("dmidecode", "-t", "17").Output()
	if err != nil {
		return nil
	}
	return parseDMIType17(string(out))
}

func parseDMIType17(output string) []model.DiscoveredDevice {
	var devices []model.DiscoveredDevice
	// Split on "Memory Device" section headers
	sections := strings.Split(output, "\nMemory Device\n")
	for i, section := range sections[1:] { // first element is preamble
		fields := parseDMIFields(section)
		sizeStr := fields["Size"]
		if sizeStr == "" || sizeStr == "No Module Installed" || sizeStr == "Not Installed" {
			continue
		}
		sizeBytes := parseDMISize(sizeStr)
		if sizeBytes == 0 {
			continue
		}

		slot := fields["Locator"]
		dimmType := fields["Type"]
		speed := fields["Speed"]
		configSpeed := fields["Configured Memory Speed"]
		if configSpeed == "" {
			configSpeed = fields["Configured Clock Speed"]
		}
		manufacturer := fields["Manufacturer"]
		partNumber := strings.TrimSpace(fields["Part Number"])
		serialNumber := strings.TrimSpace(fields["Serial Number"])
		formFactor := fields["Form Factor"]
		rank := fields["Rank"]

		vendor := normaliseDIMMVendor(manufacturer)
		stableID := fmt.Sprintf("system-dimm-%d", i)
		displayName := dimmType
		if partNumber != "" {
			displayName = partNumber
		}

		meta := map[string]interface{}{
			"size_bytes":  sizeBytes,
			"slot":        slot,
			"type":        dimmType,
			"form_factor": formFactor,
		}
		if speed != "" {
			meta["speed"] = speed
		}
		if configSpeed != "" {
			meta["configured_speed"] = configSpeed
		}
		if serialNumber != "" && serialNumber != "Not Specified" {
			meta["serial"] = serialNumber
		}
		if rank != "" {
			meta["rank"] = rank
		}

		devices = append(devices, model.DiscoveredDevice{
			StableID:       stableID,
			Platform:       "linux",
			Source:         "linux_static",
			DeviceClass:    "memory",
			DeviceSubclass: "dimm",
			Vendor:         vendor,
			Model:          partNumber,
			DisplayName:    displayName,
			Capabilities:   []string{"inventory"},
			Present:        true,
			FirstSeen:      time.Now().UTC(),
			LastSeen:       time.Now().UTC(),
			AdapterMetadata: meta,
		})
	}
	return devices
}

func parseDMIFields(section string) map[string]string {
	fields := map[string]string{}
	for _, line := range strings.Split(section, "\n") {
		// DMI fields are indented with a tab
		if !strings.HasPrefix(line, "\t") {
			continue
		}
		line = strings.TrimPrefix(line, "\t")
		idx := strings.Index(line, ": ")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+2:])
		fields[key] = val
	}
	return fields
}

func parseDMISize(s string) float64 {
	parts := strings.Fields(s)
	if len(parts) < 2 {
		return 0
	}
	val, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0
	}
	switch strings.ToUpper(parts[1]) {
	case "MB":
		return val * 1024 * 1024
	case "GB":
		return val * 1024 * 1024 * 1024
	case "TB":
		return val * 1024 * 1024 * 1024 * 1024
	}
	return 0
}

func normaliseDIMMVendor(s string) string {
	v := strings.ToLower(strings.TrimFunc(s, unicode.IsSpace))
	switch {
	case strings.Contains(v, "samsung"):
		return "samsung"
	case strings.Contains(v, "hynix") || strings.Contains(v, "sk hynix"):
		return "sk_hynix"
	case strings.Contains(v, "micron") || strings.Contains(v, "crucial"):
		return "micron"
	case strings.Contains(v, "kingston"):
		return "kingston"
	case strings.Contains(v, "corsair"):
		return "corsair"
	case strings.Contains(v, "g.skill") || strings.Contains(v, "gskill"):
		return "gskill"
	case strings.Contains(v, "teamgroup") || strings.Contains(v, "team group"):
		return "teamgroup"
	case strings.Contains(v, "patriot"):
		return "patriot"
	case v == "unknown" || v == "not specified" || v == "":
		return ""
	}
	return s
}

// dmiPlaceholders are strings that OEMs write into DMI fields when they have
// no real value to report. Treat these as absent.
var dmiPlaceholders = map[string]bool{
	"default string":          true,
	"to be filled by o.e.m.": true,
	"not applicable":          true,
	"not specified":           true,
	"none":                    true,
	"n/a":                     true,
	"unknown":                 true,
	"[s]":                     true,
}

// readDMIFile reads a DMI sysfs file and returns an empty string for known
// OEM placeholder values so they don't pollute metadata.
func readDMIFile(path string) string {
	v := readSysFile(path)
	if dmiPlaceholders[strings.ToLower(v)] {
		return ""
	}
	return v
}
