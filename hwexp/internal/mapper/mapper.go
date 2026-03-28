package mapper

import (
	"fmt"
	"hwexp/internal/model"
	"os"
	"regexp"
	"sort"
	"strconv"
	"sync"

	"gopkg.in/yaml.v3"
)

// compiledRule holds a MappingRule along with its pre-compiled regexes for performance.
type compiledRule struct {
	rule          model.MappingRule
	modelRegex    *regexp.Regexp
	stableIDRegex *regexp.Regexp
	rawNameRegex  *regexp.Regexp
}

type Engine struct {
	mu    sync.RWMutex
	rules []compiledRule
}

func NewEngine(rules []model.MappingRule) (*Engine, error) {
	e := &Engine{rules: make([]compiledRule, 0, len(rules))}
	for _, r := range rules {
		cr := compiledRule{rule: r}
		var err error
		if r.Match.ModelRegex != "" {
			cr.modelRegex, err = regexp.Compile(r.Match.ModelRegex)
			if err != nil {
				return nil, fmt.Errorf("invalid model_regex in rule %s: %w", r.ID, err)
			}
		}
		if r.Match.StableIDRegex != "" {
			cr.stableIDRegex, err = regexp.Compile(r.Match.StableIDRegex)
			if err != nil {
				return nil, fmt.Errorf("invalid stable_id_regex in rule %s: %w", r.ID, err)
			}
		}
		if r.Match.RawNameRegex != "" {
			cr.rawNameRegex, err = regexp.Compile(r.Match.RawNameRegex)
			if err != nil {
				return nil, fmt.Errorf("invalid raw_name_regex in rule %s: %w", r.ID, err)
			}
		}
		e.rules = append(e.rules, cr)
	}
	sortRules(e.rules)
	return e, nil
}

// sortRules sorts compiled rules by Priority descending so higher-priority
// rules are evaluated first in Map(). Manual rules (high priority) always
// beat auto-generated rules (priority 1).
func sortRules(rules []compiledRule) {
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].rule.Priority > rules[j].rule.Priority
	})
}

func LoadRules(path string) ([]model.MappingRule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules file: %w", err)
	}

	var wrapper struct {
		Rules []model.MappingRule `yaml:"rules"`
	}
	if err := yaml.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to parse rules file: %w", err)
	}

	return wrapper.Rules, nil
}

// AddRules compiles and appends rules that are not already present (by ID).
// Returns the IDs of rules that were actually added.
// It is safe to call concurrently with Map.
func (e *Engine) AddRules(rules []model.MappingRule) ([]string, error) {
	newEngine, err := NewEngine(rules)
	if err != nil {
		return nil, err
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	existing := make(map[string]struct{}, len(e.rules))
	for _, cr := range e.rules {
		existing[cr.rule.ID] = struct{}{}
	}

	var added []string
	for _, cr := range newEngine.rules {
		if _, dup := existing[cr.rule.ID]; !dup {
			e.rules = append(e.rules, cr)
			added = append(added, cr.rule.ID)
		}
	}
	if len(added) > 0 {
		sortRules(e.rules)
	}
	return added, nil
}

// ReloadRules reloads the mapping rules from path atomically.
// It is safe to call concurrently with Map.
func (e *Engine) ReloadRules(path string) error {
	rules, err := LoadRules(path)
	if err != nil {
		return err
	}
	newEngine, err := NewEngine(rules)
	if err != nil {
		return err
	}
	e.mu.Lock()
	e.rules = newEngine.rules
	e.mu.Unlock()
	return nil
}

// Map evaluates a raw measurement against rules and returns a normalized measurement.
func (e *Engine) Map(device model.DiscoveredDevice, raw model.RawMeasurement) (*model.NormalizedMeasurement, *model.MappingDecision) {
	e.mu.RLock()
	rules := e.rules
	e.mu.RUnlock()
	for _, cr := range rules {
		if match, matches := e.evaluateMatch(&cr, device, raw); match {
			if cr.rule.Normalize.Drop {
				return nil, &model.MappingDecision{
					MeasurementID: raw.MeasurementID,
					Decision:      "dropped",
					MappingRuleID: cr.rule.ID,
				}
			}
			return e.applyNormalization(&cr, device, raw, matches)
		}
	}

	// No rule matched
	return nil, &model.MappingDecision{
		MeasurementID: raw.MeasurementID,
		Decision:      "ignored",
	}
}

func (e *Engine) evaluateMatch(cr *compiledRule, device model.DiscoveredDevice, raw model.RawMeasurement) (bool, []string) {
	m := cr.rule.Match
	if m.Platform != "" && m.Platform != device.Platform {
		return false, nil
	}
	if m.Source != "" && m.Source != device.Source {
		return false, nil
	}
	if m.DeviceClass != "" && m.DeviceClass != device.DeviceClass {
		return false, nil
	}
	if m.DeviceSubclass != "" && m.DeviceSubclass != device.DeviceSubclass {
		return false, nil
	}
	if m.Vendor != "" && m.Vendor != device.Vendor {
		return false, nil
	}
	if m.RawName != "" && m.RawName != raw.RawName {
		return false, nil
	}
	if m.ComponentHint != "" && m.ComponentHint != raw.ComponentHint {
		return false, nil
	}
	if m.SensorHint != "" && m.SensorHint != raw.SensorHint {
		return false, nil
	}

	if cr.modelRegex != nil && !cr.modelRegex.MatchString(device.Model) {
		return false, nil
	}
	if cr.stableIDRegex != nil && !cr.stableIDRegex.MatchString(device.StableID) {
		return false, nil
	}

	var regexMatches []string
	if cr.rawNameRegex != nil {
		regexMatches = cr.rawNameRegex.FindStringSubmatch(raw.RawName)
		if regexMatches == nil {
			return false, nil
		}
	}
	return true, regexMatches
}

func (e *Engine) applyNormalization(cr *compiledRule, device model.DiscoveredDevice, raw model.RawMeasurement, regexMatches []string) (*model.NormalizedMeasurement, *model.MappingDecision) {
	n := cr.rule.Normalize

	// Apply math
	val := raw.RawValue
	if n.UnitScale != 0 {
		val *= n.UnitScale
	}
	val += n.UnitOffset

	metricFamily := expandTemplate(n.MetricFamily, device, raw, regexMatches)
	logicalName := expandTemplate(n.LogicalNameTemplate, device, raw, regexMatches)

	// Merge labels
	labels := make(map[string]string)
	labels["platform"] = device.Platform
	labels["source"] = device.Source
	labels["device_class"] = n.DeviceClass
	if labels["device_class"] == "" {
		labels["device_class"] = device.DeviceClass
	}
	labels["device_id"] = device.StableID
	labels["logical_name"] = logicalName
	labels["sensor"] = expandTemplate(n.Sensor, device, raw, regexMatches)
	labels["component"] = expandTemplate(n.Component, device, raw, regexMatches)

	for k, v := range n.Labels {
		labels[k] = expandTemplate(v, device, raw, regexMatches)
	}

	norm := &model.NormalizedMeasurement{
		StableDeviceID: device.StableID,
		LogicalName:    logicalName,
		MetricFamily:   metricFamily,
		MetricType:     n.MetricType,
		Value:          val,
		Unit:           raw.RawUnit,
		Labels:         labels,
		Quality:        raw.Quality,
		MappingRuleID:  cr.rule.ID,
		Timestamp:      raw.Timestamp,
	}

	decision := &model.MappingDecision{
		MeasurementID:  raw.MeasurementID,
		Decision:       "mapped",
		MappingRuleID:  cr.rule.ID,
		Precedence:     cr.rule.Priority,
		RawName:        raw.RawName,
		RawUnit:        raw.RawUnit,
		ConvertedValue: val,
		MetricFamily:   metricFamily,
		LogicalName:    logicalName,
	}

	return norm, decision
}

var templateTokenRE = regexp.MustCompile(`\$\{([^}]+)\}`)

func expandTemplate(t string, device model.DiscoveredDevice, raw model.RawMeasurement, regexMatches []string) string {
	if t == "" {
		return t
	}

	return templateTokenRE.ReplaceAllStringFunc(t, func(token string) string {
		m := templateTokenRE.FindStringSubmatch(token)
		if len(m) != 2 {
			return token
		}
		key := m[1]

		if key == "logical_device_name" {
			return device.LogicalDeviceName
		}
		if key == "stable_device_id" {
			return device.StableID
		}

		// Regex capture placeholders: ${0}, ${1}, ...
		if idx, err := strconv.Atoi(key); err == nil {
			if idx >= 0 && idx < len(regexMatches) {
				return regexMatches[idx]
			}
			return ""
		}

		// Metadata placeholders: ${zone}, ${state}, ${mc}, ...
		if raw.Metadata != nil {
			if v, ok := raw.Metadata[key]; ok {
				return v
			}
		}

		return ""
	})
}
