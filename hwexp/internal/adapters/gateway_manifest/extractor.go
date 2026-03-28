package gateway_manifest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// ExtractJSONValue performs basic JQ-style extraction from a JSON body.
// Supported patterns:
// ".field"
// ".list | length"
// ".list[0].field"
func ExtractJSONValue(data []byte, path string) (float64, error) {
	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return 0, err
	}

	// Simple JQ-style path evaluator
	return evaluatePath(obj, path)
}

func evaluatePath(obj interface{}, path string) (float64, error) {
	path = strings.TrimSpace(path)
	if path == "" || path == "." {
		return toFloat(obj)
	}

	// Handle pipe for length
	if strings.Contains(path, "|") {
		parts := strings.Split(path, "|")
		baseObj, err := navigatePath(obj, strings.TrimSpace(parts[0]))
		if err != nil {
			return 0, err
		}
		op := strings.TrimSpace(parts[1])
		if op == "length" {
			if l, ok := baseObj.([]interface{}); ok {
				return float64(len(l)), nil
			}
			if m, ok := baseObj.(map[string]interface{}); ok {
				return float64(len(m)), nil
			}
			return 0, fmt.Errorf("length operator applied to non-collection")
		}
	}

	finalObj, err := navigatePath(obj, path)
	if err != nil {
		return 0, err
	}
	return toFloat(finalObj)
}

func navigatePath(obj interface{}, path string) (interface{}, error) {
	parts := strings.Split(strings.TrimPrefix(path, "."), ".")
	curr := obj

	for _, part := range parts {
		if part == "" {
			continue
		}

		// Handle array index: field[0]
		if strings.Contains(part, "[") && strings.Contains(part, "]") {
			openIdx := strings.Index(part, "[")
			closeIdx := strings.Index(part, "]")
			fieldName := part[:openIdx]
			indexStr := part[openIdx+1 : closeIdx]
			index, _ := strconv.Atoi(indexStr)

			if fieldName != "" {
				m, ok := curr.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("path error: %s is not a map", fieldName)
				}
				curr = m[fieldName]
			}

			l, ok := curr.([]interface{})
			if !ok {
				return nil, fmt.Errorf("path error: not an array")
			}
			if index < 0 || index >= len(l) {
				return nil, fmt.Errorf("path error: index out of bounds")
			}
			curr = l[index]
		} else {
			m, ok := curr.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("path error: %s is not a map", part)
			}
			curr = m[part]
		}
	}
	return curr, nil
}

func toFloat(obj interface{}) (float64, error) {
	switch v := obj.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to float", obj)
	}
}

// ExtractPrometheusValue parses Prometheus text format and returns the first value matching the name.
func ExtractPrometheusValue(data []byte, name string) (float64, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// Handle metrics with labels: metric_name{labels} value
		metricPart := fields[0]
		if idx := strings.Index(metricPart, "{"); idx > 0 {
			metricPart = metricPart[:idx]
		}

		if metricPart == name {
			return strconv.ParseFloat(fields[1], 64)
		}
	}
	return 0, fmt.Errorf("metric %s not found", name)
}
