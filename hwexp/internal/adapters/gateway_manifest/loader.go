package gateway_manifest

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	"hwexp/internal/model"
)

func LoadManifests(projectDir, localDir string) ([]model.Manifest, error) {
	manifests := make(map[string]model.Manifest)

	// 1. Load project manifests
	projectManifests := make(map[string]model.Manifest)
	if projectDir != "" {
		if err := loadFromDir(projectDir, projectManifests); err != nil {
			return nil, err
		}
	}

	// 2. Load local overrides
	localManifests := make(map[string]model.Manifest)
	if localDir != "" {
		if err := loadFromDir(localDir, localManifests); err != nil {
			return nil, err
		}
	}

	// 3. Merge
	for id, projectM := range projectManifests {
		if localM, ok := localManifests[id]; ok {
			manifests[id] = mergeManifests(projectM, localM)
		} else {
			manifests[id] = projectM
		}
	}
	// Add local-only manifests
	for id, localM := range localManifests {
		if _, ok := manifests[id]; !ok {
			manifests[id] = localM
		}
	}

	var result []model.Manifest
	for _, m := range manifests {
		if m.Enabled {
			result = append(result, resolveManifest(m))
		}
	}
	return result, nil
}

func mergeManifests(p, l model.Manifest) model.Manifest {
	res := p
	if l.DisplayName != "" { res.DisplayName = l.DisplayName }
	// Keep project enabled flag by default; local overrides can still provide
	// a full manifest if they need to explicitly disable.
	// (bool zero-value cannot distinguish "unset" from "false" here.)
	
	// Merge Health
	if l.Health.Endpoint != "" { res.Health.Endpoint = l.Health.Endpoint }
	if l.Health.ExpectStatus != 0 { res.Health.ExpectStatus = l.Health.ExpectStatus }
	if l.Health.TimeoutS != 0 { res.Health.TimeoutS = l.Health.TimeoutS }

	// Merge Connection
	if l.Connection.Host != "" { res.Connection.Host = l.Connection.Host }
	if l.Connection.Port != 0 { res.Connection.Port = l.Connection.Port }
	if l.Connection.Auth.Type != "" { res.Connection.Auth.Type = l.Connection.Auth.Type }
	if l.Connection.Auth.TokenEnv != "" { res.Connection.Auth.TokenEnv = l.Connection.Auth.TokenEnv }

	// Merge Discovery
	if l.Discovery != nil {
		if res.Discovery == nil {
			res.Discovery = l.Discovery
		} else {
			if l.Discovery.Type != "" { res.Discovery.Type = l.Discovery.Type }
			if l.Discovery.Endpoint != "" { res.Discovery.Endpoint = l.Discovery.Endpoint }
			if l.Discovery.ActivityField != "" { res.Discovery.ActivityField = l.Discovery.ActivityField }
			if l.Discovery.BackendPortField != "" { res.Discovery.BackendPortField = l.Discovery.BackendPortField }
		}
	}

	// Merge Correlation
	if l.Correlation != nil {
		if res.Correlation == nil {
			res.Correlation = l.Correlation
		} else {
			if l.Correlation.DeviceClass != "" { res.Correlation.DeviceClass = l.Correlation.DeviceClass }
			if l.Correlation.PCISlot != "" { res.Correlation.PCISlot = l.Correlation.PCISlot }
		}
	}

	// Merge Metrics list by ID
	metricMap := make(map[string]model.MetricConfig)
	for _, pm := range p.Metrics {
		metricMap[pm.ID] = pm
	}
	for _, lm := range l.Metrics {
		if pm, ok := metricMap[lm.ID]; ok {
			// Merge metric fields
			if lm.Endpoint != "" { pm.Endpoint = lm.Endpoint }
			if lm.Extract != "" { pm.Extract = lm.Extract }
			if lm.PrometheusName != "" { pm.PrometheusName = lm.PrometheusName }
			if lm.Unit != "" { pm.Unit = lm.Unit }
			if lm.PollIntervalS != 0 { pm.PollIntervalS = lm.PollIntervalS }
			if lm.SourceFormat != "" { pm.SourceFormat = lm.SourceFormat }
			metricMap[lm.ID] = pm
		} else {
			metricMap[lm.ID] = lm
		}
	}
	res.Metrics = nil
	for _, m := range metricMap {
		res.Metrics = append(res.Metrics, m)
	}

	return res
}

func loadFromDir(dir string, manifests map[string]model.Manifest) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}

		var m model.Manifest
		if err := yaml.Unmarshal(data, &m); err != nil {
			continue
		}

		if m.ID == "" {
			continue
		}

		// Shallow merge: existing entries are overwritten by later ones (local wins)
		manifests[m.ID] = m
	}
	return nil
}

func resolveManifest(m model.Manifest) model.Manifest {
	m.DisplayName = ResolveVariables(m.DisplayName)
	m.Health.Endpoint = ResolveVariables(m.Health.Endpoint)
	m.Connection.Host = ResolveVariables(m.Connection.Host)
	if m.Discovery != nil {
		m.Discovery.Endpoint = ResolveVariables(m.Discovery.Endpoint)
	}
	if m.Correlation != nil {
		m.Correlation.PCISlot = ResolveVariables(m.Correlation.PCISlot)
	}

	for i := range m.Metrics {
		m.Metrics[i].Endpoint = ResolveVariables(m.Metrics[i].Endpoint)
		m.Metrics[i].Extract = ResolveVariables(m.Metrics[i].Extract)
		m.Metrics[i].PrometheusName = ResolveVariables(m.Metrics[i].PrometheusName)
	}
	return m
}
