// File: backup_estimate.go
// Deskripsi: Versi paling ringkas untuk estimasi backup dan pengecekan ruang disk.
// Author: Hadiyatna Muflihun
// Tanggal: 16 Oktober 2025

package backup

import (
	"fmt"
	"math"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/compress"

	"github.com/shirou/gopsutil/v3/disk"
)

// --- Constants for Estimation Logic ---
const (
	// sqlDumpMultiplier memperkirakan ukuran SQL dump mentah menjadi ~1.35x lebih besar dari ukuran data biner
	// karena adanya statement CREATE/INSERT dan formatting teks.
	sqlDumpMultiplier = 0.85

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
func (s *Service) CheckDiskSpaceForBackup(outputDir string, dbNames []string) error {

	if len(dbNames) == 0 {
		return fmt.Errorf("tidak ada database yang dipilih untuk diestimasi")
	}

	// Pre-allocate estimates slice to avoid repeated allocations when jumlah db diketahui.
	estimates := make([]structs.BackupSizeEstimate, 0, len(dbNames))
	var totalEstimatedSize float64 // Gunakan float64 untuk presisi perhitungan

	// --- 1. Mulai Estimasi Ukuran (Logika digabung di sini) ---
	for _, dbName := range dbNames {
		detail, exists := s.DatabaseDetail[dbName]
		if !exists || detail.SizeBytes <= 0 {
			continue
		}

		// Hitung ukuran dump SQL mentah
		finalSize := float64(detail.SizeBytes) * sqlDumpMultiplier
		// ukuran dump SQL mentah dibulatkan ke atas saat disimpan sebagai uint64
		estimatedSQLDumpSize := uint64(math.Ceil(float64(detail.SizeBytes) * sqlDumpMultiplier))

		// Dapatkan rasio kompresi dengan helper untuk mengurangi cabang nested
		ratio := s.getCompressionRatio()
		if s.EstimateOptions.CompressionEnabled {
			finalSize *= ratio
		}

		// Tambahkan overhead enkripsi
		if s.EstimateOptions.EncryptionEnabled {
			finalSize *= encryptionOverheadMultiplier
		}

		// Bulatkan ke atas untuk memastikan kita tidak meremehkan kebutuhan ruang
		estimatedFinal := uint64(math.Ceil(finalSize))

		estimates = append(estimates, structs.BackupSizeEstimate{
			DatabaseName:         dbName,
			OriginalSize:         detail.SizeBytes,
			EstimatedSQLDumpSize: estimatedSQLDumpSize,
			EstimatedFinalSize:   estimatedFinal,
			CompressionRatio:     ratio,
			CompressionEnabled:   s.EstimateOptions.CompressionEnabled,
			EncryptionEnabled:    s.EstimateOptions.EncryptionEnabled,
		})
		totalEstimatedSize += float64(estimatedFinal)
	}

	// Tambahkan overhead untuk mode "combined"
	if s.EstimateOptions.BackupMode == "combined" && len(estimates) > 1 {
		totalEstimatedSize *= combinedBackupOverheadMultiplier
	}

	// --- 2. Cek Ruang Disk ---
	usage, err := disk.Usage(outputDir)
	if err != nil {
		return fmt.Errorf("gagal memeriksa ruang disk di %s: %w", outputDir, err)
	}

	// --- 3. Finalisasi Hasil ---
	estimatedBackupSize := uint64(totalEstimatedSize)
	safetyFactor := 1.0 + (s.EstimateOptions.SafetyMarginPct / 100.0)
	requiredWithMargin := uint64(float64(estimatedBackupSize) * safetyFactor)

	s.DiskSpaceCheckResult = &structs.DiskSpaceCheckResult{
		OutputDirectory:         outputDir,
		EstimatedBackupSize:     estimatedBackupSize,
		RequiredWithMargin:      requiredWithMargin,
		AvailableDiskSpace:      usage.Free,
		HasEnoughSpace:          usage.Free >= requiredWithMargin,
		DatabaseEstimates:       estimates,
		DatabasesWithoutDetails: len(dbNames) - len(estimates),
	}

	return nil
}

// getCompressionRatio mengembalikan rasio kompresi yang berlaku berdasarkan opsi.
// Nilai yang dikembalikan adalah 1.0 jika kompresi tidak diaktifkan atau tidak ada
// rasio spesifik tersedia.
func (s *Service) getCompressionRatio() float64 {
	if !s.EstimateOptions.CompressionEnabled || s.EstimateOptions.CompressionType == compress.CompressionNone {
		return 1.0
	}

	if levelMap, ok := compress.CompressionRatios[s.EstimateOptions.CompressionType]; ok {
		if r, ok := levelMap[s.EstimateOptions.CompressionLevel]; ok {
			return r
		}
		if r, ok := levelMap[compress.LevelDefault]; ok {
			return r
		}
	}

	// Fallback ke rasio default ketika tipe kompresi dikenal tapi level tidak ditemukan,
	// atau ketika tipe kompresi tidak punya entry di map (misalnya plugin eksternal)
	return defaultCompressionRatio
}
