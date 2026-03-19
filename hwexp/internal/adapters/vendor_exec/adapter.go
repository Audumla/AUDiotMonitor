package vendor_exec

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"hwexp/internal/model"
)

type Adapter struct {
	scriptsDir string
}

func NewAdapter(dir string) *Adapter {
	if dir == "" {
		dir = "/etc/hwexp/custom.d"
	}
	return &Adapter{scriptsDir: dir}
}

type DiscoveryResult struct {
	Devices []model.DiscoveredDevice `json:"devices"`
}

func (a *Adapter) Discover(ctx context.Context) ([]model.DiscoveredDevice, error) {
	files, err := os.ReadDir(a.scriptsDir)
	if err != nil {
		if os.IsNotExist(err) { return nil, nil }
		return nil, err
	}

	var allDevices []model.DiscoveredDevice
	for _, f := range files {
		if f.IsDir() { continue }
		
		path := filepath.Join(a.scriptsDir, f.Name())
		// Run with --discover flag
		cmd := exec.CommandContext(ctx, path, "--discover")
		out, err := cmd.Output()
		if err != nil {
			continue
		}

		var res DiscoveryResult
		if err := json.Unmarshal(out, &res); err == nil {
			allDevices = append(allDevices, res.Devices...)
		}
	}
	return allDevices, nil
}

func (a *Adapter) Poll(ctx context.Context) ([]model.RawMeasurement, error) {
	files, err := os.ReadDir(a.scriptsDir)
	if err != nil {
		if os.IsNotExist(err) { return nil, nil }
		return nil, err
	}

	var allMetrics []model.RawMeasurement
	for _, f := range files {
		if f.IsDir() { continue }
		
		path := filepath.Join(a.scriptsDir, f.Name())
		cmd := exec.CommandContext(ctx, path)
		out, err := cmd.Output()
		if err != nil {
			continue
		}

		var metrics []model.RawMeasurement
		if err := json.Unmarshal(out, &metrics); err == nil {
			allMetrics = append(allMetrics, metrics...)
		}
	}
	return allMetrics, nil
}
