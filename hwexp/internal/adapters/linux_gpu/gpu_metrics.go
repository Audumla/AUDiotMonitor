package linux_gpu

// gpu_metrics reads the binary gpu_metrics sysfs file exposed by the amdgpu
// driver. This is the preferred source for GPU utilisation on RDNA/RDNA2/RDNA3
// cards because it provides an atomic snapshot that the SMU firmware updates
// every ~1 ms, rather than a stale percentage that may only update when the
// display driver polls it.
//
// Reference: Linux kernel include/uapi/linux/amdgpu_drm.h and
// drivers/gpu/drm/amd/pm/swsmu/inc/amdgpu_smu.h
//
// Binary layout (all fields little-endian):
//
//	Header (4 bytes)
//	  [0:1]  structure_size   uint16
//	  [2]    format_revision  uint8   – 1, 2, or 3
//	  [3]    content_revision uint8
//
//	v1.x (Arcturus / CDNA server cards):
//	  [4:5]  temperature_edge     uint16
//	  [6:7]  temperature_hotspot  uint16
//	  [8:9]  temperature_mem      uint16
//	  [10:11] temperature_vrgfx   uint16
//	  [12:13] temperature_vrsoc   uint16
//	  [14:15] temperature_vrmem   uint16
//	  [16:17] average_gfx_activity  uint16  ← percent 0–100
//	  [18:19] average_umc_activity  uint16  ← percent 0–100
//
//	v2.x (RDNA2: Navi 21/22/23/24, RX 6000 series):
//	  [4:5]  temperature_gfx      uint16
//	  [6:7]  temperature_soc      uint16
//	  [8:23] temperature_core[8]  [8]uint16
//	  [24:27] temperature_l3[2]   [2]uint16
//	  [28:29] average_gfx_activity uint16  ← percent 0–100
//	  [30:31] average_umc_activity uint16  ← percent 0–100
//	  [32:33] average_mm_activity  uint16  ← VCN/media engine
//
//	v3.x (RDNA3: Navi 31/32/33, RX 7000 series):
//	  [4:5]  temperature_gfx       uint16
//	  [6:7]  temperature_soc       uint16
//	  [8:9]  temperature_vrsoc     uint16
//	  [10:11] current_socket_power uint16
//	  [12:13] average_gfx_activity uint16  ← percent 0–100
//	  [14:15] average_umc_activity uint16  ← percent 0–100

import (
	"encoding/binary"
	"os"
	"path/filepath"
)

// gpuMetrics holds the utilisation values extracted from the gpu_metrics file.
type gpuMetrics struct {
	GFXActivity uint16 // GPU compute engine busy, 0–100 %
	UMCActivity uint16 // Memory-controller busy, 0–100 %
	MMActivity  uint16 // Media (VCN/UVD) engine busy, 0–100 % (v2.x only)
	Valid       bool
}

// readGPUMetrics parses /sys/class/drm/cardN/device/gpu_metrics.
// sysfsDevPath is the device sysfs directory, e.g. /sys/class/drm/card0/device.
// Returns an invalid (zero) result if the file is absent or unrecognised.
func readGPUMetrics(sysfsDevPath string) gpuMetrics {
	path := filepath.Join(sysfsDevPath, "gpu_metrics")
	data, err := os.ReadFile(path)
	if err != nil || len(data) < 4 {
		return gpuMetrics{}
	}

	// Header layout: [0:1] size, [2] format_revision, [3] content_revision
	formatRev := data[2]

	switch formatRev {
	case 1:
		// v1.x – Arcturus / CDNA (server).  average_gfx_activity at offset 16.
		if len(data) < 20 {
			return gpuMetrics{}
		}
		return gpuMetrics{
			GFXActivity: binary.LittleEndian.Uint16(data[16:18]),
			UMCActivity: binary.LittleEndian.Uint16(data[18:20]),
			Valid:        true,
		}

	case 2:
		// v2.x – RDNA2.  average_gfx_activity at offset 28.
		if len(data) < 34 {
			return gpuMetrics{}
		}
		return gpuMetrics{
			GFXActivity: binary.LittleEndian.Uint16(data[28:30]),
			UMCActivity: binary.LittleEndian.Uint16(data[30:32]),
			MMActivity:  binary.LittleEndian.Uint16(data[32:34]),
			Valid:        true,
		}

	case 3:
		// v3.x – RDNA3.  average_gfx_activity at offset 12.
		if len(data) < 16 {
			return gpuMetrics{}
		}
		return gpuMetrics{
			GFXActivity: binary.LittleEndian.Uint16(data[12:14]),
			UMCActivity: binary.LittleEndian.Uint16(data[14:16]),
			Valid:        true,
		}

	default:
		return gpuMetrics{}
	}
}
