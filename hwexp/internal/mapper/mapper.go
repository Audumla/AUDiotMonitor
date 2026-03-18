package mapper

import (
	"fmt"
	"hwexp/internal/model"
	"os"
	"regexp"
	"strings"

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
	// Note: A robust implementation would sort e.rules by Priority descending here.
	return e, nil
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

// Map evaluates a raw measurement against rules and returns a normalized measurement.
func (e *Engine) Map(device model.DiscoveredDevice, raw model.RawMeasurement) (*model.NormalizedMeasurement, *model.MappingDecision) {
	for _, cr := range e.rules {
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
	if m.Platform != "" && m.Platform != device.Platform { return false, nil }
	if m.Source != "" && m.Source != device.Source { return false, nil }
	if m.DeviceClass != "" && m.DeviceClass != device.DeviceClass { return false, nil }
	if m.Vendor != "" && m.Vendor != device.Vendor { return false, nil }
	if m.ComponentHint != "" && m.ComponentHint != raw.ComponentHint { return false, nil }

	if cr.modelRegex != nil && !cr.modelRegex.MatchString(device.Model) { return false, nil }
	if cr.stableIDRegex != nil && !cr.stableIDRegex.MatchString(device.StableID) { return false, nil }

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

	// Expand logical name template. Simple replacement for ${1}, ${2} etc.
	// For production, regex.Expand is safer.
	logicalName := n.LogicalNameTemplate
	logicalName = strings.ReplaceAll(logicalName, "${logical_device_name}", device.LogicalDeviceName)
	for i, matchStr := range regexMatches {
		placeholder := fmt.Sprintf("${%d}", i)
		logicalName = strings.ReplaceAll(logicalName, placeholder, matchStr)
	}

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
	labels["sensor"] = n.Sensor
	labels["component"] = n.Component

	for k, v := range n.Labels {
		labels[k] = v
	}

	norm := &model.NormalizedMeasurement{
		StableDeviceID: device.StableID,
		LogicalName:    logicalName,
		MetricFamily:   n.MetricFamily,
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
		MetricFamily:   n.MetricFamily,
		LogicalName:    logicalName,
	}

	return norm, decision
}
