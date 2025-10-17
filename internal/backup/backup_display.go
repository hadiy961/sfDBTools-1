package backup

import (
	"fmt"
	"path/filepath"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/ui"
)

// DisplaySummaryTable adalah "controller" yang mengatur tampilan summary.
func (s *Service) DisplaySummaryTable(summary *BackupSummary) {
	ui.Headers("BACKUP SUMMARY") // <-- Menggunakan ui.PrintHeader

	s.displayGeneralInfo(summary)
	s.displayServerInfo(summary)
	s.displayDBStats(summary)
	s.displayOutputInfo(summary)
	s.displayConfig(summary)
	s.displaySuccessfulDBs(summary)
	s.displayDatabaseDetails(summary)
	s.displayFailedDBs(summary)
	s.displayErrors(summary)
}

// Masing-masing fungsi di bawah ini sekarang hanya bertanggung jawab untuk SATU tabel.
func (s *Service) displayGeneralInfo(summary *BackupSummary) {
	ui.PrintSubHeader("Informasi Umum")
	data := [][]string{
		{"Backup ID", summary.BackupID},
		{"Status", ui.GetStatusIcon(summary.Status) + " " + summary.Status}, // <-- Menggunakan ui.GetStatusIcon
		{"Mode", summary.BackupMode},
		{"Waktu Mulai", summary.StartTime.Format(displayTimeFormat)},
		{"Waktu Selesai", summary.EndTime.Format(displayTimeFormat)},
		{"Durasi", summary.Duration},
	}
	ui.FormatTable([]string{"Property", "Value"}, data)
}

func (s *Service) displayDBStats(summary *BackupSummary) {
	ui.PrintSubHeader("Statistik Database")
	data := [][]string{
		{"Total Database", fmt.Sprintf("%d", summary.DatabaseStats.TotalDatabases)},
		{"Berhasil", fmt.Sprintf("✅ %d", summary.DatabaseStats.SuccessfulBackups)},
		{"Gagal", fmt.Sprintf("❌ %d", summary.DatabaseStats.FailedBackups)},
	}
	if summary.DatabaseStats.ExcludedDatabases > 0 {
		data = append(data, []string{"Dikecualikan", fmt.Sprintf("⚠️ %d", summary.DatabaseStats.ExcludedDatabases)})
	}
	ui.FormatTable([]string{"Metric", "Count"}, data)
}

// displayServerInfo menampilkan informasi server database (tanpa password)
func (s *Service) displayServerInfo(summary *BackupSummary) {
	// Jika tidak ada info server yang tersedia, lewati
	if summary.ServerInfo.Host == "" && summary.ServerInfo.User == "" && summary.ServerInfo.Port == 0 {
		return
	}
	ui.PrintSubHeader("Informasi Server")
	data := [][]string{
		{"Host", summary.ServerInfo.Host},
		{"Port", fmt.Sprintf("%d", summary.ServerInfo.Port)},
		{"User", summary.ServerInfo.User},
	}
	if summary.ServerInfo.Database != "" {
		data = append(data, []string{"Database", summary.ServerInfo.Database})
	}
	if summary.ServerInfo.Config != "" {
		data = append(data, []string{"Config", summary.ServerInfo.Config})
	}
	if summary.ServerInfo.Version != "" {
		data = append(data, []string{"Version", summary.ServerInfo.Version})
	}
	ui.FormatTable([]string{"Property", "Value"}, data)
}

func (s *Service) displayOutputInfo(summary *BackupSummary) {
	ui.PrintSubHeader("Output Files")
	data := [][]string{
		{"Direktori Output", summary.OutputInfo.OutputDirectory},
		{"Total File", fmt.Sprintf("%d", summary.OutputInfo.TotalFiles)},
		{"Total Ukuran", summary.OutputInfo.TotalSizeHuman},
	}
	ui.FormatTable([]string{"Property", "Value"}, data)
}

func (s *Service) displayConfig(summary *BackupSummary) {
	ui.PrintSubHeader("Konfigurasi Backup")
	var compressionStatus, compressionDetail = "❌ Disabled", "-"
	if summary.BackupConfig.CompressionEnabled {
		compressionStatus = "✅ Enabled"
		compressionDetail = fmt.Sprintf("%s (level: %s)", summary.BackupConfig.CompressionType, summary.BackupConfig.CompressionLevel)
	}
	var encryptionStatus, encryptionDetail = "❌ Disabled", "-"
	if summary.BackupConfig.EncryptionEnabled {
		encryptionStatus = "✅ Enabled"
		encryptionDetail = "AES-256-GCM"
	}
	var cleanupStatus, cleanupDetail = "❌ Disabled", "-"
	if summary.BackupConfig.CleanupEnabled {
		cleanupStatus = "✅ Enabled"
		cleanupDetail = fmt.Sprintf("%d hari", summary.BackupConfig.RetentionDays)
	}
	data := [][]string{
		{"Kompresi", compressionStatus, compressionDetail},
		{"Enkripsi", encryptionStatus, encryptionDetail},
		{"Auto Cleanup", cleanupStatus, cleanupDetail},
	}
	ui.FormatTable([]string{"Fitur", "Status", "Detail"}, data)
}

func (s *Service) displaySuccessfulDBs(summary *BackupSummary) {
	if len(summary.SuccessfulDatabases) == 0 {
		return
	}
	ui.PrintSubHeader("Database Berhasil")

	// Cek apakah ada data ukuran database asli
	hasOriginalSize := false
	for _, db := range summary.SuccessfulDatabases {
		if db.OriginalDBSize > 0 {
			hasOriginalSize = true
			break
		}
	}

	var data [][]string
	var headers []string

	if hasOriginalSize {
		headers = []string{"Database", "File Output", "DB Size", "Backup Size", "Ratio", "Durasi"}
		for _, db := range summary.SuccessfulDatabases {
			dbSizeStr := db.OriginalDBSizeHuman
			if dbSizeStr == "" {
				dbSizeStr = "-"
			}

			// Hitung compression ratio
			ratioStr := "-"
			if db.CompressionRatio > 0 {
				ratioStr = fmt.Sprintf("%.2f%%", db.CompressionRatio*100)
			}

			data = append(data, []string{
				db.DatabaseName,
				filepath.Base(db.OutputFile),
				dbSizeStr,
				db.FileSizeHuman,
				ratioStr,
				db.Duration,
			})
		}
	} else {
		headers = []string{"Database", "File Output", "Ukuran", "Durasi"}
		for _, db := range summary.SuccessfulDatabases {
			data = append(data, []string{
				db.DatabaseName,
				filepath.Base(db.OutputFile),
				db.FileSizeHuman,
				db.Duration,
			})
		}
	}

	ui.FormatTable(headers, data)
}

func (s *Service) displayDatabaseDetails(summary *BackupSummary) {
	if len(summary.DatabaseDetails) == 0 {
		return
	}
	ui.PrintSubHeader("Detail Database")
	var data [][]string
	for _, db := range summary.SuccessfulDatabases {
		if detail, exists := summary.DatabaseDetails[db.DatabaseName]; exists {
			data = append(data, []string{
				detail.DatabaseName, detail.SizeHuman, fmt.Sprintf("%d", detail.TableCount),
				fmt.Sprintf("%d", detail.ViewCount), fmt.Sprintf("%d", detail.ProcedureCount),
				fmt.Sprintf("%d", detail.FunctionCount), fmt.Sprintf("%d", detail.UserGrantCount),
			})
		}
	}
	if len(data) > 0 {
		ui.FormatTable([]string{"Database", "DB Size", "Tables", "Views", "Procs", "Funcs", "Users"}, data)
	}
}

func (s *Service) displayFailedDBs(summary *BackupSummary) {
	if len(summary.FailedDatabases) == 0 {
		return
	}
	ui.PrintSubHeader("Database Gagal")
	var data [][]string
	for _, db := range summary.FailedDatabases {
		data = append(data, []string{db.DatabaseName, db.Error})
	}
	ui.FormatTable([]string{"Database", "Error"}, data)
}

func (s *Service) displayErrors(summary *BackupSummary) {
	if len(summary.Errors) == 0 {
		return
	}
	ui.PrintSubHeader("Error Umum")
	for _, err := range summary.Errors {
		ui.PrintColoredLine("   • "+err, ui.ColorYellow)
	}
	fmt.Println()
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
			ui.FormatBytesInt64(est.OriginalSize),
			ui.FormatBytes(est.EstimatedSQLDumpSize),
			ui.FormatBytes(est.EstimatedFinalSize),
			compressionInfo,
		})
	}

	// Tambahkan baris total
	data = append(data, []string{
		"",
		fmt.Sprintf("TOTAL (%d DB)", len(estimates)),
		ui.FormatBytesInt64(totalOriginal),
		ui.FormatBytes(totalSQLDump),
		ui.FormatBytes(totalFinal),
		"",
	})

	ui.FormatTable(headers, data)
}

func (s *Service) RingkasanDiskCheck(dbFiltered []string) error {
	// Tampilkan ringkasan
	ui.PrintSubHeader("RINGKASAN PENGECEKAN RUANG DISK")
	// Informasi estimasi

	s.Logger.Infof("Total database: %d", len(dbFiltered))
	if s.DiskSpaceCheckResult.DatabasesWithoutDetails > 0 {
		s.Logger.Warnf("Database tanpa detail ukuran: %d (tidak termasuk dalam estimasi)", s.DiskSpaceCheckResult.DatabasesWithoutDetails)
	}
	s.Logger.Infof("Estimasi ukuran backup: %s", ui.FormatBytes(s.DiskSpaceCheckResult.EstimatedBackupSize))
	s.Logger.Infof("Dengan safety margin (%.0f%%): %s", s.EstimateOptions.SafetyMarginPct, ui.FormatBytes(s.DiskSpaceCheckResult.RequiredWithMargin))

	s.Logger.Info("Informasi Disk:")
	s.Logger.Infof("  Tersedia: %s", ui.FormatBytes(s.DiskSpaceCheckResult.AvailableDiskSpace))

	// Tampilkan status dan pesan
	if s.DiskSpaceCheckResult.HasEnoughSpace {
		sisaSetelahBackup := s.DiskSpaceCheckResult.AvailableDiskSpace - s.DiskSpaceCheckResult.RequiredWithMargin
		s.Logger.Info("✓ STATUS: RUANG DISK MENCUKUPI")
		s.Logger.Infof("  Sisa ruang setelah backup: %s", ui.FormatBytes(sisaSetelahBackup))
	} else {
		shortage := s.getShortage(s.DiskSpaceCheckResult)
		s.Logger.Error("✗ STATUS: RUANG DISK TIDAK MENCUKUPI")
		s.Logger.Errorf("  Kekurangan ruang: %s", ui.FormatBytes(shortage))
		s.Logger.Error("")
		s.Logger.Error("TINDAKAN:")
		s.Logger.Error("  1. Bersihkan file yang tidak diperlukan")
		s.Logger.Error("  2. Gunakan direktori output di partisi lain")
		s.Logger.Error("  3. Aktifkan kompresi dengan level lebih tinggi")
		s.Logger.Error("  4. Kurangi jumlah database yang di-backup")

		return fmt.Errorf("ruang disk tidak mencukupi untuk backup (kekurangan: %s)", ui.FormatBytes(shortage))
	}

	return nil
}
