// File: backup_disk_check.go
// Deskripsi: Fungsi-fungsi untuk pengecekan ruang disk sebelum backup
// Author: AI Assistant
// Tanggal: 16 Oktober 2025
// Last Modified: 16 Oktober 2025

package backup

import (
	"context"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/compress"
)

// checkDiskSpaceBeforeBackup melakukan pengecekan ruang disk sebelum memulai backup
// Mengembalikan map estimasi per database untuk dibandingkan dengan hasil actual
func (s *Service) checkDiskSpaceBeforeBackup(
	ctx context.Context,
	config BackupConfig,
	dbFiltered []string,
	backupMode string,
) (map[string]uint64, error) {
	// Tentukan compression settings
	compressionEnabled := config.CompressionRequired
	compType := compress.CompressionType(config.CompressionType)
	compLevel := compress.CompressionLevel(s.BackupOptions.Compression.Level)
	encryptionEnabled := s.BackupOptions.Encryption.Enabled

	// Safety margin default 15% (bisa dikonfigurasi)
	safetyMarginPct := 10.0

	// Buat options untuk estimasi
	s.EstimateOptions = &structs.EstimateOptions{
		CompressionEnabled: compressionEnabled,
		CompressionType:    compType,
		CompressionLevel:   compLevel,
		EncryptionEnabled:  encryptionEnabled,
		BackupMode:         backupMode,
		SafetyMarginPct:    safetyMarginPct,
	}

	// Lakukan pengecekan ruang disk dengan estimasi
	err := s.CheckDiskSpaceForBackup(config.OutputDir, dbFiltered)

	if err != nil {
		s.Logger.Warnf("Gagal melakukan pengecekan ruang disk: %v", err)
		s.Logger.Warn("Melanjutkan backup tanpa pengecekan ruang disk...")
		return make(map[string]uint64), nil // Tidak fatal, hanya warning
	}

	// Buat map estimasi per database untuk digunakan nanti
	estimatesMap := make(map[string]uint64)
	for _, est := range s.DiskSpaceCheckResult.DatabaseEstimates {
		estimatesMap[est.DatabaseName] = est.EstimatedFinalSize
	}

	// Tampilkan detail estimasi per database jika ada
	if len(s.DiskSpaceCheckResult.DatabaseEstimates) > 0 {
		s.displayDatabaseEstimatesTable(s.DiskSpaceCheckResult.DatabaseEstimates)
	}

	if err := s.ringkasanDiskCheck(dbFiltered); err != nil {
		// ringkasanDiskCheck sudah mengembalikan error yang menjelaskan kekurangan ruang
		return nil, err
	}
	return estimatesMap, nil
}
