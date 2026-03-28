package gateway_manifest

import (
	"os"
	"regexp"
)

var varRegex = regexp.MustCompile(`\$\{([^}:]+)(?::-(.*))?\}`)

// ResolveVariables replaces ${VAR:-default} patterns in a string with environment variables.
func ResolveVariables(input string) string {
	return varRegex.ReplaceAllStringFunc(input, func(m string) string {
		match := varRegex.FindStringSubmatch(m)
		if len(match) < 2 {
			return m
		}
		name := match[1]
		defaultValue := ""
		if len(match) > 2 {
			defaultValue = match[2]
		}

		val := os.Getenv(name)
		if val == "" {
			return defaultValue
		}
		return val
	})
}

// ResolveInMap recursively resolves variables in a map (useful for metadata).
func ResolveInMap(m map[string]interface{}) {
	for k, v := range m {
		if s, ok := v.(string); ok {
			m[k] = ResolveVariables(s)
		} else if nm, ok := v.(map[string]interface{}); ok {
			ResolveInMap(nm)
		}
	}
}
