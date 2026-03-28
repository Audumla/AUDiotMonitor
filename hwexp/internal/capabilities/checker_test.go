package capabilities

import (
	"errors"
	"testing"
)

type fakeProvider struct {
	reqs []Requirement
}

func (f fakeProvider) Requirements() []Requirement {
	return f.reqs
}

type fakeLogger struct {
	calls []map[string]interface{}
}

func (f *fakeLogger) Info(event, message string, details map[string]interface{}) {
	f.calls = append(f.calls, details)
}

func TestCheckRequirements_MixedAvailability(t *testing.T) {
	providers := []Provider{
		fakeProvider{reqs: []Requirement{
			{Name: "smartctl", Description: "storage health"},
			{Name: "dmidecode", Description: "memory inventory", Optional: true},
		}},
	}
	logger := &fakeLogger{}
	lookPath := func(file string) (string, error) {
		if file == "smartctl" {
			return "/usr/bin/smartctl", nil
		}
		return "", errors.New("not found")
	}

	status := CheckRequirements(providers, lookPath, logger)

	if !status["smartctl"] {
		t.Fatalf("expected smartctl to be available")
	}
	if status["dmidecode"] {
		t.Fatalf("expected dmidecode to be unavailable")
	}
	if len(logger.calls) != 1 {
		t.Fatalf("expected one missing dependency log, got %d", len(logger.calls))
	}
}

func TestCheckRequirements_DeduplicatesDependencyStatus(t *testing.T) {
	providers := []Provider{
		fakeProvider{reqs: []Requirement{{Name: "smartctl"}}},
		fakeProvider{reqs: []Requirement{{Name: "smartctl"}}},
	}
	lookPath := func(file string) (string, error) {
		return "/usr/bin/" + file, nil
	}

	status := CheckRequirements(providers, lookPath, nil)

	if len(status) != 1 {
		t.Fatalf("expected one dependency entry, got %d", len(status))
	}
	if !status["smartctl"] {
		t.Fatalf("expected smartctl to be available")
	}
}
