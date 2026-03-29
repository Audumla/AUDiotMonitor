package join

import (
	"testing"

	"hwexp/internal/model"
)

func TestIndexDeviceByPCISlot(t *testing.T) {
	index := make(map[string]model.DiscoveredDevice)
	device := model.DiscoveredDevice{
		StableID: "gpu-1",
		RawIdentifiers: map[string]string{
			"pci_slot": "0000:65:00.0",
		},
	}

	IndexDeviceByPCISlot(index, device)

	got, ok := index["0000:65:00.0"]
	if !ok {
		t.Fatalf("expected device to be indexed by pci_slot")
	}
	if got.StableID != "gpu-1" {
		t.Fatalf("expected indexed device stable id gpu-1, got %q", got.StableID)
	}
}

func TestEnrichNormalizedMeasurement_InjectsCorrelationLabels(t *testing.T) {
	norm := &model.NormalizedMeasurement{
		Labels: map[string]string{
			"source": "gateway_manifest",
		},
	}
	raw := model.RawMeasurement{
		Metadata: map[string]string{
			"correlation_pci_slot": "0000:65:00.0",
		},
	}
	deviceIndex := map[string]model.DiscoveredDevice{
		"0000:65:00.0": {
			StableID:    "gpu-1",
			Vendor:      "amd",
			Model:       "7900xtx",
			DisplayName: "GPU 0",
		},
	}

	EnrichNormalizedMeasurement(norm, raw, deviceIndex)

	if norm.Labels["device_id"] != "gpu-1" {
		t.Fatalf("expected device_id label gpu-1, got %q", norm.Labels["device_id"])
	}
	if norm.Labels["vendor"] != "amd" {
		t.Fatalf("expected vendor label amd, got %q", norm.Labels["vendor"])
	}
	if norm.Labels["model"] != "7900xtx" {
		t.Fatalf("expected model label 7900xtx, got %q", norm.Labels["model"])
	}
	if norm.Labels["display_name"] != "GPU 0" {
		t.Fatalf("expected display_name label GPU 0, got %q", norm.Labels["display_name"])
	}
}

func TestEnrichNormalizedMeasurement_AddsModelNameFromMetadata(t *testing.T) {
	norm := &model.NormalizedMeasurement{}
	raw := model.RawMeasurement{
		Metadata: map[string]string{
			"model_name": "llama3.1-70b",
		},
	}

	EnrichNormalizedMeasurement(norm, raw, map[string]model.DiscoveredDevice{})

	if norm.Labels["model_name"] != "llama3.1-70b" {
		t.Fatalf("expected model_name label llama3.1-70b, got %q", norm.Labels["model_name"])
	}
}
