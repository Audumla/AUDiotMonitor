package vendor_exec

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"hwexp/internal/model"
)

type Adapter struct {
	scriptsDir   string
	pollTimeout  time.Duration
	sourceFormat string // "json" or "prometheus"
}

func NewAdapter(dir string, timeout time.Duration, format string) *Adapter {
	if dir == "" {
		dir = "/etc/hwexp/custom.d"
	}
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	if format == "" {
		format = "json"
	}
	return &Adapter{
		scriptsDir:   dir,
		pollTimeout:  timeout,
		sourceFormat: format,
	}
}

type DiscoveryResult struct {
	Devices []model.DiscoveredDevice `json:"devices"`
}

func (a *Adapter) Discover(ctx context.Context) ([]model.DiscoveredDevice, error) {
	files, err := os.ReadDir(a.scriptsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var allDevices []model.DiscoveredDevice
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		path := filepath.Join(a.scriptsDir, f.Name())
		
		// Use a sub-context with timeout for each script
		subCtx, cancel := context.WithTimeout(ctx, a.pollTimeout)
		cmd := exec.CommandContext(subCtx, path, "--discover")
		out, err := cmd.Output()
		cancel()

		if err != nil {
			log.Printf("vendor_exec: discovery script %s failed: %v", f.Name(), err)
			continue
		}

		var res DiscoveryResult
		if err := json.Unmarshal(out, &res); err != nil {
			log.Printf("vendor_exec: failed to parse discovery output from %s: %v", f.Name(), err)
			continue
		}
		allDevices = append(allDevices, res.Devices...)
	}
	return allDevices, nil
}

func (a *Adapter) Poll(ctx context.Context) ([]model.RawMeasurement, error) {
	files, err := os.ReadDir(a.scriptsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var allMetrics []model.RawMeasurement
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		path := filepath.Join(a.scriptsDir, f.Name())
		
		subCtx, cancel := context.WithTimeout(ctx, a.pollTimeout)
		cmd := exec.CommandContext(subCtx, path)
		out, err := cmd.Output()
		cancel()

		if err != nil {
			log.Printf("vendor_exec: poll script %s failed: %v", f.Name(), err)
			continue
		}

		var metrics []model.RawMeasurement
		if a.sourceFormat == "prometheus" {
			metrics, err = parsePrometheusMetrics(out, f.Name())
		} else {
			err = json.Unmarshal(out, &metrics)
		}

		if err != nil {
			log.Printf("vendor_exec: failed to parse poll output from %s (format=%s): %v", f.Name(), a.sourceFormat, err)
			continue
		}
		allMetrics = append(allMetrics, metrics...)
	}
	return allMetrics, nil
}

// parsePrometheusMetrics is a simple parser for Prometheus text format.
// It maps the first word to RawName and the second to RawValue.
// It expects each line to be: metric_name value
func parsePrometheusMetrics(data []byte, scriptName string) ([]model.RawMeasurement, error) {
	var results []model.RawMeasurement
	scanner := bufio.NewScanner(bytes.NewReader(data))
	now := time.Now().UTC()

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		name := fields[0]
		val, err := strconv.ParseFloat(fields[1], 64)
		if err != nil {
			continue
		}

		results = append(results, model.RawMeasurement{
			MeasurementID:  fmt.Sprintf("vendor_exec:%s:%s", scriptName, name),
			Source:         "linux_vendor_exec",
			RawName:        name,
			RawValue:       val,
			Timestamp:      now,
			Quality:        "good",
		})
	}
	return results, nil
}
