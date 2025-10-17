// File: backup_estimate.go
// Deskripsi: Versi paling ringkas untuk estimasi backup dan pengecekan ruang disk.
// Author: Hadiyatna Muflihun
// Tanggal: 16 Oktober 2025

package fs

import (
	"fmt"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/compress"

	"github.com/dustin/go-humanize"
	"github.com/shirou/gopsutil/v3/disk"
)

// --- Constants for Estimation Logic ---
const (
	// sqlDumpMultiplier memperkirakan ukuran SQL dump mentah menjadi ~1.35x lebih besar dari ukuran data biner
	// karena adanya statement CREATE/INSERT dan formatting teks.
	sqlDumpMultiplier = 1.35

	// encryptionOverheadMultiplier menambahkan overhead ~2% untuk header dan padding enkripsi.
	encryptionOverheadMultiplier = 1.02

	// combinedBackupOverheadMultiplier menambahkan overhead ~1% untuk backup mode "combined"
	// yang menampung beberapa database dalam satu file.
	combinedBackupOverheadMultiplier = 1.01

	// defaultCompressionRatio adalah fallback jika tipe kompresi tidak ditemukan.
	defaultCompressionRatio = 0.15
)

// pgzip memiliki rasio yang sama dengan gzip standar.
var _ = func() interface{} {
	compress.CompressionRatios[compress.CompressionPgzip] = compress.CompressionRatios[compress.CompressionGzip]
	return nil
}()

// --- Core Function ---

// CheckDiskSpaceForBackup adalah satu-satunya fungsi yang dibutuhkan.
// Logika estimasi dan pengecekan digabung untuk alur yang lebih linier.
func CheckDiskSpaceForBackup(
	outputDir string,
	dbDetails map[string]structs.DatabaseDetail,
	dbNames []string,
	opts structs.EstimateOptions,
) (*structs.DiskSpaceCheckResult, error) {

	if len(dbNames) == 0 {
		return nil, fmt.Errorf("tidak ada database yang dipilih untuk diestimasi")
	}

	var estimates []structs.BackupSizeEstimate
	var totalEstimatedSize float64 // Gunakan float64 untuk presisi perhitungan

	// --- 1. Mulai Estimasi Ukuran (Logika digabung di sini) ---
	for _, dbName := range dbNames {
		detail, exists := dbDetails[dbName]
		if !exists || detail.SizeBytes <= 0 {
			continue
		}

		// Hitung ukuran dump SQL mentah
		finalSize := float64(detail.SizeBytes) * sqlDumpMultiplier
		estimatedSQLDumpSize := uint64(finalSize)

		// Dapatkan rasio kompresi (logika `getCompressionRatio` di-inline)
		ratio := 1.0
		if opts.CompressionEnabled {
			if levelMap, ok := compress.CompressionRatios[opts.CompressionType]; ok {
				if r, ok := levelMap[opts.CompressionLevel]; ok {
					ratio = r
				} else if r, ok := levelMap[compress.LevelDefault]; ok {
					ratio = r // Fallback ke level default
				}
			} else if opts.CompressionType != compress.CompressionNone {
				ratio = defaultCompressionRatio // Fallback ke rasio default
			}
		}

		finalSize *= ratio

		// Tambahkan overhead enkripsi
		if opts.EncryptionEnabled {
			finalSize *= encryptionOverheadMultiplier
		}

		estimates = append(estimates, structs.BackupSizeEstimate{
			DatabaseName:         dbName,
			OriginalSize:         detail.SizeBytes,
			EstimatedSQLDumpSize: estimatedSQLDumpSize,
			EstimatedFinalSize:   uint64(finalSize),
			CompressionRatio:     ratio,
			CompressionEnabled:   opts.CompressionEnabled,
			EncryptionEnabled:    opts.EncryptionEnabled,
		})
		totalEstimatedSize += finalSize
	}

	// Tambahkan overhead untuk mode "combined"
	if opts.BackupMode == "combined" && len(estimates) > 1 {
		totalEstimatedSize *= combinedBackupOverheadMultiplier
	}

	// --- 2. Cek Ruang Disk ---
	usage, err := disk.Usage(outputDir)
	if err != nil {
		return nil, fmt.Errorf("gagal memeriksa ruang disk di %s: %w", outputDir, err)
	}

	// --- 3. Finalisasi Hasil ---
	estimatedBackupSize := uint64(totalEstimatedSize)
	safetyFactor := 1.0 + (opts.SafetyMarginPct / 100.0)
	requiredWithMargin := uint64(float64(estimatedBackupSize) * safetyFactor)

	result := &structs.DiskSpaceCheckResult{
		OutputDirectory:         outputDir,
		EstimatedBackupSize:     estimatedBackupSize,
		RequiredWithMargin:      requiredWithMargin,
		AvailableDiskSpace:      usage.Free,
		HasEnoughSpace:          usage.Free >= requiredWithMargin,
		DatabaseEstimates:       estimates,
		DatabasesWithoutDetails: len(dbNames) - len(estimates),
	}

	return result, nil
}

// GetShortage menghitung kekurangan ruang disk jika tidak cukup.
func GetShortage(r *structs.DiskSpaceCheckResult) uint64 {
	if r.HasEnoughSpace || r.AvailableDiskSpace >= r.RequiredWithMargin {
		return 0
	}
	return r.RequiredWithMargin - r.AvailableDiskSpace
}

// GetSummaryMessage mengembalikan pesan ringkasan yang mudah dibaca.
func GetSummaryMessage(r *structs.DiskSpaceCheckResult) string {
	if r.HasEnoughSpace {
		return fmt.Sprintf(
			"✓ Ruang disk cukup:\n"+
				"  Estimasi Ukuran Backup: %s\n"+
				"  Dibutuhkan (dgn margin): %s\n"+
				"  Ruang Tersedia: %s\n"+
				"  Sisa Setelah Backup: %s",
			humanize.Bytes(r.EstimatedBackupSize),
			humanize.Bytes(r.RequiredWithMargin),
			humanize.Bytes(r.AvailableDiskSpace),
			humanize.Bytes(r.AvailableDiskSpace-r.RequiredWithMargin),
		)
	}

	return fmt.Sprintf(
		"✗ Ruang disk TIDAK cukup:\n"+
			"  Estimasi Ukuran Backup: %s\n"+
			"  Dibutuhkan (dgn margin): %s\n"+
			"  Ruang Tersedia: %s\n"+
			"  Kekurangan: %s",
		humanize.Bytes(r.EstimatedBackupSize),
		humanize.Bytes(r.RequiredWithMargin),
		humanize.Bytes(r.AvailableDiskSpace),
		humanize.Bytes(GetShortage(r)),
	)
}
