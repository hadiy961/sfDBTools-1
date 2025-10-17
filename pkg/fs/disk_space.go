// File: disk_space.go
// Deskripsi: Package untuk pengecekan dan estimasi ruang disk
// Author: AI Assistant
// Tanggal: 16 Oktober 2025
// Last Modified: 16 Oktober 2025

package fs

import (
	"fmt"
	"math"

	// Library eksternal yang matang untuk informasi disk
	"github.com/shirou/gopsutil/v3/disk"
	// Library eksternal yang matang untuk format bytes
)

// DiskSpaceInfo menyimpan informasi ruang disk
type DiskSpaceInfo struct {
	Total     uint64  // Total ruang disk dalam bytes
	Free      uint64  // Ruang kosong dalam bytes
	Available uint64  // Ruang yang tersedia untuk non-root user dalam bytes
	Used      uint64  // Ruang yang terpakai dalam bytes
	UsedPct   float64 // Persentase penggunaan
}

// GetDiskSpace mengambil informasi ruang disk untuk path tertentu
func GetDiskSpace(path string) (*DiskSpaceInfo, error) {
	// Menggunakan gopsutil untuk portabilitas
	usage, err := disk.Usage(path)
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan informasi disk untuk %s: %w", path, err)
	}

	return &DiskSpaceInfo{
		Total:     usage.Total,
		Free:      usage.Free,
		Available: usage.Free, // Menggunakan Free untuk ketersediaan
		Used:      usage.Used,
		UsedPct:   usage.UsedPercent,
	}, nil
}

// HasEnoughSpace memeriksa apakah ada cukup ruang disk untuk ukuran yang diminta
// dengan menambahkan safety margin (default 10%)
func HasEnoughSpace(path string, requiredBytes uint64, safetyMarginPct float64) (bool, *DiskSpaceInfo, error) {
	// Memastikan margin dalam batas 0% - 100%
	if safetyMarginPct < 0 {
		safetyMarginPct = 0
	}
	safetyMarginPct = math.Min(safetyMarginPct, 100.0)

	diskInfo, err := GetDiskSpace(path)
	if err != nil {
		return false, nil, err
	}

	// Hitung kebutuhan termasuk safety margin, dengan pembulatan ke atas (Ceil)
	safetyFactor := 1.0 + (safetyMarginPct / 100.0)
	requiredWithMargin := uint64(math.Ceil(float64(requiredBytes) * safetyFactor))

	return diskInfo.Available >= requiredWithMargin, diskInfo, nil
}
