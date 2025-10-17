// File: backup_disk_check.go
// Deskripsi: Fungsi-fungsi untuk pengecekan ruang disk sebelum backup
// Author: AI Assistant
// Tanggal: 16 Oktober 2025
// Last Modified: 16 Oktober 2025

package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/fs"
	"sfDBTools/pkg/ui"
)

// checkDiskSpaceBeforeBackup melakukan pengecekan ruang disk sebelum memulai backup
// Mengembalikan map estimasi per database untuk dibandingkan dengan hasil actual
func (s *Service) checkDiskSpaceBeforeBackup(
	ctx context.Context,
	config BackupConfig,
	dbFiltered []string,
	databaseDetails map[string]structs.DatabaseDetail,
	backupMode string,
) (map[string]uint64, error) {
	ui.PrintSubHeader("Pengecekan Ruang Disk Sebelum Backup")
	// Tentukan compression settings
	compressionEnabled := config.CompressionRequired
	compType := compress.CompressionType(config.CompressionType)
	compLevel := compress.CompressionLevel(s.BackupOptions.Compression.Level)
	encryptionEnabled := s.BackupOptions.Encryption.Enabled

	// Safety margin default 15% (bisa dikonfigurasi)
	safetyMarginPct := 15.0

	s.Logger.Infof("Direktori output: %s", config.OutputDir)

	// Buat options untuk estimasi
	opts := structs.EstimateOptions{
		CompressionEnabled: compressionEnabled,
		CompressionType:    compType,
		CompressionLevel:   compLevel,
		EncryptionEnabled:  encryptionEnabled,
		BackupMode:         backupMode,
		SafetyMarginPct:    safetyMarginPct,
	}

	// Lakukan pengecekan
	result, err := fs.CheckDiskSpaceForBackup(
		config.OutputDir,
		databaseDetails,
		dbFiltered,
		opts,
	)

	if err != nil {
		s.Logger.Warnf("Gagal melakukan pengecekan ruang disk: %v", err)
		s.Logger.Warn("Melanjutkan backup tanpa pengecekan ruang disk...")
		return make(map[string]uint64), nil // Tidak fatal, hanya warning
	}

	// Buat map estimasi per database untuk digunakan nanti
	estimatesMap := make(map[string]uint64)
	for _, est := range result.DatabaseEstimates {
		estimatesMap[est.DatabaseName] = est.EstimatedFinalSize
	}

	// Tampilkan detail estimasi per database jika ada
	if len(result.DatabaseEstimates) > 0 {
		s.displayDatabaseEstimatesTable(result.DatabaseEstimates)
	}

	// Tampilkan ringkasan
	ui.PrintSubHeader("RINGKASAN PENGECEKAN RUANG DISK")
	// Informasi estimasi
	s.Logger.Infof("Total database: %d", len(dbFiltered))
	if result.DatabasesWithoutDetails > 0 {
		s.Logger.Warnf("Database tanpa detail ukuran: %d (tidak termasuk dalam estimasi)", result.DatabasesWithoutDetails)
	}
	s.Logger.Infof("Estimasi ukuran backup: %s", fs.FormatBytes(result.EstimatedBackupSize))
	s.Logger.Infof("Dengan safety margin (%.0f%%): %s", safetyMarginPct, fs.FormatBytes(result.RequiredWithMargin))

	s.Logger.Info("Informasi Disk:")
	s.Logger.Infof("  Tersedia: %s", fs.FormatBytes(result.AvailableDiskSpace))

	// Tampilkan status dan pesan
	if result.HasEnoughSpace {
		sisaSetelahBackup := result.AvailableDiskSpace - result.RequiredWithMargin
		s.Logger.Info("✓ STATUS: RUANG DISK MENCUKUPI")
		s.Logger.Infof("  Sisa ruang setelah backup: %s", fs.FormatBytes(sisaSetelahBackup))
	} else {
		shortage := fs.GetShortage(result)
		s.Logger.Error("✗ STATUS: RUANG DISK TIDAK MENCUKUPI")
		s.Logger.Errorf("  Kekurangan ruang: %s", fs.FormatBytes(shortage))
		s.Logger.Error("")
		s.Logger.Error("TINDAKAN:")
		s.Logger.Error("  1. Bersihkan file yang tidak diperlukan")
		s.Logger.Error("  2. Gunakan direktori output di partisi lain")
		s.Logger.Error("  3. Aktifkan kompresi dengan level lebih tinggi")
		s.Logger.Error("  4. Kurangi jumlah database yang di-backup")

		return nil, fmt.Errorf("ruang disk tidak mencukupi untuk backup (kekurangan: %s)", fs.FormatBytes(shortage))
	}
	return estimatesMap, nil
}

// displayDatabaseEstimatesTable menampilkan tabel estimasi ukuran per database
func (s *Service) displayDatabaseEstimatesTable(estimates []structs.BackupSizeEstimate) {
	if len(estimates) == 0 {
		return
	}
	ui.PrintSubHeader("ESTIMASI UKURAN PER DATABASE")

	headers := []string{"No", "Database", "Ukuran Asli", "Estimasi SQL Dump", "Estimasi Final", "Compression Ratio"}
	var data [][]string

	var totalOriginal int64
	var totalSQLDump uint64
	var totalFinal uint64

	for i, est := range estimates {
		totalOriginal += est.OriginalSize
		totalSQLDump += est.EstimatedSQLDumpSize
		totalFinal += est.EstimatedFinalSize

		compressionInfo := "No compression"
		if est.CompressionEnabled {
			compressionInfo = fmt.Sprintf("%.1f%%", est.CompressionRatio*100)
		}

		data = append(data, []string{
			fmt.Sprintf("%d", i+1),
			est.DatabaseName,
			fs.FormatBytesInt64(est.OriginalSize),
			fs.FormatBytes(est.EstimatedSQLDumpSize),
			fs.FormatBytes(est.EstimatedFinalSize),
			compressionInfo,
		})
	}

	// Tambahkan baris total
	data = append(data, []string{
		"",
		fmt.Sprintf("TOTAL (%d DB)", len(estimates)),
		fs.FormatBytesInt64(totalOriginal),
		fs.FormatBytes(totalSQLDump),
		fs.FormatBytes(totalFinal),
		"",
	})

	ui.FormatTable(headers, data)
}
