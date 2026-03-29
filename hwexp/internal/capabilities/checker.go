package capabilities

import "os/exec"

// EventLogger is the minimal logger contract required for requirement checks.
type EventLogger interface {
	Info(event, message string, details map[string]interface{})
}

// LookPathFunc resolves binaries in PATH.
type LookPathFunc func(file string) (string, error)

// CheckRequirements validates runtime dependencies declared by providers.
// It returns a map keyed by dependency name with availability status.
func CheckRequirements(providers []Provider, lookPath LookPathFunc, l EventLogger) map[string]bool {
	if lookPath == nil {
		lookPath = exec.LookPath
	}

	status := map[string]bool{}
	for _, provider := range providers {
		for _, req := range provider.Requirements() {
			if _, seen := status[req.Name]; !seen {
				status[req.Name] = false
			}
			if _, err := lookPath(req.Name); err == nil {
				status[req.Name] = true
				continue
			}
			if l != nil {
				l.Info("capability_missing", "Adapter runtime dependency missing", map[string]interface{}{
					"dependency":  req.Name,
					"description": req.Description,
					"optional":    req.Optional,
				})
			}
		}
	}

	return status
}
