package join

import "hwexp/internal/model"

// IndexDeviceByPCISlot registers a device in the correlation index when a PCI
// slot identifier is available.
func IndexDeviceByPCISlot(index map[string]model.DiscoveredDevice, device model.DiscoveredDevice) {
	if index == nil {
		return
	}
	if slot, ok := device.RawIdentifiers["pci_slot"]; ok && slot != "" {
		index[slot] = device
	}
}

// EnrichNormalizedMeasurement injects correlation labels into a normalized
// measurement using metadata hints from the raw sample.
func EnrichNormalizedMeasurement(
	norm *model.NormalizedMeasurement,
	raw model.RawMeasurement,
	deviceIndex map[string]model.DiscoveredDevice,
) {
	if norm == nil || raw.Metadata == nil {
		return
	}

	if slot, ok := raw.Metadata["correlation_pci_slot"]; ok && slot != "" {
		if targetDev, found := deviceIndex[slot]; found {
			ensureLabels(norm)
			norm.Labels["device_id"] = targetDev.StableID
			norm.Labels["vendor"] = targetDev.Vendor
			norm.Labels["model"] = targetDev.Model
			if targetDev.DisplayName != "" {
				norm.Labels["display_name"] = targetDev.DisplayName
			}
		}
	}

	// model_name supports two-tier source discovery (for example gateway backends).
	if modelName, ok := raw.Metadata["model_name"]; ok {
		ensureLabels(norm)
		norm.Labels["model_name"] = modelName
	}
}

func ensureLabels(norm *model.NormalizedMeasurement) {
	if norm.Labels == nil {
		norm.Labels = make(map[string]string)
	}
}
