package pcidb

import (
	"bufio"
	"os"
	"strings"
	"sync"
)

// Standard paths where pci.ids may be installed.
var searchPaths = []string{
	"/usr/share/hwdata/pci.ids",   // openSUSE, Fedora, RHEL
	"/usr/share/misc/pci.ids",     // Debian/Ubuntu, Alpine (pciutils)
	"/usr/share/pci.ids",
	"/etc/pci.ids",
}

type db struct {
	vendors map[string]string            // hex4 → vendor name
	devices map[string]map[string]string // vendorHex → deviceHex → device name
}

var (
	once     sync.Once
	globalDB *db
)

func load() *db {
	once.Do(func() {
		for _, p := range searchPaths {
			if d := parse(p); d != nil {
				globalDB = d
				return
			}
		}
		globalDB = &db{
			vendors: map[string]string{},
			devices: map[string]map[string]string{},
		}
	})
	return globalDB
}

func parse(path string) *db {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	d := &db{
		vendors: map[string]string{},
		devices: map[string]map[string]string{},
	}

	var curVendor string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		// Two-tab lines are subsystem entries — skip
		if strings.HasPrefix(line, "\t\t") {
			continue
		}
		// One-tab line: device entry under current vendor
		if strings.HasPrefix(line, "\t") {
			if curVendor == "" {
				continue
			}
			line = line[1:]
			if id, name, ok := parseEntry(line); ok {
				if d.devices[curVendor] == nil {
					d.devices[curVendor] = map[string]string{}
				}
				d.devices[curVendor][id] = name
			}
			continue
		}
		// Vendor entry
		if id, name, ok := parseEntry(line); ok {
			d.vendors[id] = name
			curVendor = id
		}
	}
	return d
}

// parseEntry extracts "id  name" from a pci.ids line.
// The id is always exactly 4 hex characters; the name follows with any whitespace.
func parseEntry(line string) (id, name string, ok bool) {
	if len(line) < 6 {
		return
	}
	raw := strings.ToLower(line[:4])
	for _, c := range raw {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			return
		}
	}
	n := strings.TrimSpace(line[4:])
	if n == "" {
		return
	}
	return raw, n, true
}

// Lookup returns the vendor and device names for the given 4-char hex IDs.
// Returns empty strings when not found or when the database is unavailable.
func Lookup(vendorHex, deviceHex string) (vendorName, deviceName string) {
	d := load()
	v := strings.ToLower(strings.TrimSpace(vendorHex))
	dv := strings.ToLower(strings.TrimSpace(deviceHex))
	vendorName = d.vendors[v]
	if devMap, ok := d.devices[v]; ok {
		deviceName = devMap[dv]
	}
	return
}

// Available returns true if a pci.ids file was found and loaded.
func Available() bool {
	return len(load().vendors) > 0
}
