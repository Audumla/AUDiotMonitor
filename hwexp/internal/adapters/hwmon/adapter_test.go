package hwmon

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestAdapter_Discover(t *testing.T) {
	err := os.MkdirAll(filepath.Join("tmp", "sys", "class", "hwmon"), 0755)
	if err != nil {
		t.Fatalf("Failed to create tmp dir: %v", err)
	}
	tmpDir := filepath.Join("tmp", "sys", "class", "hwmon")
	defer os.RemoveAll("tmp")

	// Create a mock hwmon0 device
	hwmon0 := filepath.Join(tmpDir, "hwmon0")
	os.MkdirAll(hwmon0, 0755)
	os.WriteFile(filepath.Join(hwmon0, "name"), []byte("amdgpu\n"), 0644)
	
	// Create device directory and uevent for stable ID
	os.MkdirAll(filepath.Join(hwmon0, "device"), 0755)
	os.WriteFile(filepath.Join(hwmon0, "device", "uevent"), []byte("PCI_SLOT_NAME=0000:0b:00.0\n"), 0644)

	// Create some sensor files
	os.WriteFile(filepath.Join(hwmon0, "temp1_input"), []byte("54000\n"), 0644)
	os.WriteFile(filepath.Join(hwmon0, "fan1_input"), []byte("1200\n"), 0644)

	adapter := NewAdapter(tmpDir)
	devices, err := adapter.Discover(context.Background())
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	if len(devices) != 1 {
		t.Errorf("Expected 1 device, got %d", len(devices))
	}

	dev := devices[0]
	if dev.Driver != "amdgpu" {
		t.Errorf("Expected driver amdgpu, got %s", dev.Driver)
	}
	if dev.StableID != "pci-0000:0b:00.0" {
		t.Errorf("Expected stable ID pci-0000:0b:00.0, got %s", dev.StableID)
	}
}

func TestAdapter_Poll(t *testing.T) {
	os.MkdirAll(filepath.Join("tmp_poll", "sys", "class", "hwmon"), 0755)
	tmpDir := filepath.Join("tmp_poll", "sys", "class", "hwmon")
	defer os.RemoveAll("tmp_poll")

	hwmon0 := filepath.Join(tmpDir, "hwmon0")
	os.MkdirAll(hwmon0, 0755)
	os.WriteFile(filepath.Join(hwmon0, "name"), []byte("coretemp\n"), 0644)
	os.WriteFile(filepath.Join(hwmon0, "temp1_input"), []byte("42000\n"), 0644)

	adapter := NewAdapter(tmpDir)
	measurements, err := adapter.Poll(context.Background())
	if err != nil {
		t.Fatalf("Poll failed: %v", err)
	}

	if len(measurements) != 1 {
		t.Fatalf("Expected 1 measurement, got %d", len(measurements))
	}

	m := measurements[0]
	if m.RawValue != 42000 {
		t.Errorf("Expected raw value 42000, got %f", m.RawValue)
	}
	if m.RawUnit != "millidegree_celsius" {
		t.Errorf("Expected unit millidegree_celsius, got %s", m.RawUnit)
	}
}
