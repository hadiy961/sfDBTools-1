// File : internal/backup/backup_summary.go
// Deskripsi : Fungsi untuk membuat, menyimpan, dan menampilkan summary backup.
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025 (Refactored for clarity and best practices)

package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/ui" // <-- Menggunakan package UI Anda
	"time"

	"github.com/dustin/go-humanize"
)

const (
	// backupIDTimeFormat adalah format waktu yang digunakan untuk generate Backup ID.
	backupIDTimeFormat = "20060102_150405"
	// displayTimeFormat adalah format waktu yang mudah dibaca untuk ditampilkan di summary.
	displayTimeFormat = "2006-01-02 15:04:05"
)

// CreateBackupSummary membuat summary backup yang komprehensif.
// Fungsi ini menggabungkan pembuatan summary dengan atau tanpa detail database untuk mengurangi duplikasi kode.
// Jika detail database tidak tersedia, pass `nil` pada parameter `databaseDetails`.
func (s *Service) CreateBackupSummary(
	backupMode string,
	dbFiltered []string,
	successfulDBs []DatabaseBackupInfo,
	failedDBs []FailedDatabaseInfo,
	startTime time.Time,
	errors []string,
	databaseDetails map[string]structs.DatabaseDetail,
) *BackupSummary {
	endTime := time.Now()

	// Tentukan status backup berdasarkan hasil
	var status string
	switch {
	case len(failedDBs) == 0 && len(successfulDBs) > 0:
		status = "success"
	case len(failedDBs) > 0 && len(successfulDBs) > 0:
		status = "partial"
	case len(failedDBs) > 0 && len(successfulDBs) == 0:
		status = "failed"
	default:
		status = "empty" // Tidak ada database yang diproses
	}

	summary := &BackupSummary{
		BackupID:            fmt.Sprintf("backup_%s", startTime.Format(backupIDTimeFormat)),
		Timestamp:           startTime,
		BackupMode:          backupMode,
		Status:              status,
		Duration:            s.formatDuration(endTime.Sub(startTime)),
		StartTime:           startTime,
		EndTime:             endTime,
		DatabaseStats:       s.buildDatabaseStats(dbFiltered, successfulDBs, failedDBs),
		OutputInfo:          s.collectOutputInfo(successfulDBs),
		BackupConfig:        s.buildBackupConfigSummary(),
		SuccessfulDatabases: successfulDBs,
		FailedDatabases:     failedDBs,
		DatabaseDetails:     databaseDetails,
		ServerInfo: ServerConnectionInfo{
			Host:     s.DBConfigInfo.ServerDBConnection.Host,
			Port:     s.DBConfigInfo.ServerDBConnection.Port,
			User:     s.DBConfigInfo.ServerDBConnection.User,
			Database: s.DBConfigInfo.ServerDBConnection.Database,
			Config:   s.DBConfigInfo.ConfigName,
		},
		Errors: errors,
	}

	return summary
}

// buildDatabaseStats membuat statistik database.
func (s *Service) buildDatabaseStats(dbFiltered []string, successfulDBs []DatabaseBackupInfo, failedDBs []FailedDatabaseInfo) DatabaseSummaryStats {
	stats := DatabaseSummaryStats{
		TotalDatabases:    len(dbFiltered),
		SuccessfulBackups: len(successfulDBs),
		FailedBackups:     len(failedDBs),
	}
	if s.FilterInfo != nil {
		stats.ExcludedDatabases = s.FilterInfo.ExcludedDatabases
		stats.SystemDatabases = s.FilterInfo.SystemDatabases
		stats.FilteredDatabases = s.FilterInfo.IncludedDatabases
	}
	return stats
}

// buildBackupConfigSummary membuat ringkasan konfigurasi dari state Service.
func (s *Service) buildBackupConfigSummary() BackupConfigSummary {
	cfg := BackupConfigSummary{
		CompressionEnabled: s.BackupOptions.Compression.Enabled,
		EncryptionEnabled:  s.BackupOptions.Encryption.Enabled,
		CleanupEnabled:     s.BackupOptions.Cleanup.Enabled,
	}

	if cfg.CompressionEnabled {
		cfg.CompressionType = s.BackupOptions.Compression.Type
		cfg.CompressionLevel = s.BackupOptions.Compression.Level
	}
	if s.BackupOptions.DBList != "" {
		cfg.DBListFile = s.BackupOptions.DBList
	}
	if cfg.CleanupEnabled {
		cfg.RetentionDays = s.BackupOptions.Cleanup.RetentionDays
	}
	return cfg
}

// collectOutputInfo mengumpulkan informasi file output dari backup yang berhasil.
func (s *Service) collectOutputInfo(successfulDBs []DatabaseBackupInfo) OutputSummaryInfo {
	var files []SummaryFileInfo
	var totalSize int64

	for _, db := range successfulDBs {
		fileInfo := SummaryFileInfo{
			FileName:     filepath.Base(db.OutputFile),
			FilePath:     db.OutputFile,
			Size:         db.FileSize,
			SizeHuman:    db.FileSizeHuman,
			DatabaseName: db.DatabaseName,
			CreatedAt:    time.Now(),
		}
		files = append(files, fileInfo)
		totalSize += db.FileSize
	}

	return OutputSummaryInfo{
		OutputDirectory: s.BackupOptions.OutputDirectory,
		TotalFiles:      len(files),
		TotalSize:       totalSize,
		TotalSizeHuman:  humanize.Bytes(uint64(totalSize)),
		Files:           files,
	}
}

// SaveSummaryToJSON menyimpan summary ke file JSON.
func (s *Service) SaveSummaryToJSON(summary *BackupSummary) error {
	baseDirectory := s.Config.Backup.Output.BaseDirectory
	summaryDir := filepath.Join(baseDirectory, "summaries")
	if err := os.MkdirAll(summaryDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori summary: %w", err)
	}

	summaryFileName := fmt.Sprintf("%s.json", summary.BackupID)
	summaryPath := filepath.Join(summaryDir, summaryFileName)

	jsonData, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("gagal marshal summary ke JSON: %w", err)
	}

	if err := os.WriteFile(summaryPath, jsonData, 0644); err != nil {
		return fmt.Errorf("gagal menulis file summary: %w", err)
	}

	s.Logger.Infof("Summary backup disimpan ke: %s", summaryPath)
	return nil
}

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

	// Cek apakah ada estimasi data
	hasEstimates := false
	for _, db := range summary.SuccessfulDatabases {
		if db.EstimatedSize > 0 {
			hasEstimates = true
			break
		}
	}

	var data [][]string
	var headers []string

	if hasEstimates {
		headers = []string{"Database", "File Output", "Estimated", "Actual", "Accuracy", "Durasi"}
		for _, db := range summary.SuccessfulDatabases {
			accuracyStr := "-"
			if db.AccuracyPercentage > 0 {
				accuracyStr = fmt.Sprintf("%.1f%%", db.AccuracyPercentage)
			}
			estimatedStr := db.EstimatedSizeHuman
			if estimatedStr == "" {
				estimatedStr = "-"
			}

			data = append(data, []string{
				db.DatabaseName,
				filepath.Base(db.OutputFile),
				estimatedStr,
				db.FileSizeHuman,
				accuracyStr,
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

// formatDuration memformat durasi menjadi string yang mudah dibaca.
// Fungsi ini tetap private karena hanya relevan untuk package ini.
func (s *Service) formatDuration(duration time.Duration) string {
	switch {
	case duration < time.Minute:
		return fmt.Sprintf("%.1f detik", duration.Seconds())
	case duration < time.Hour:
		return fmt.Sprintf("%.1f menit", duration.Minutes())
	default:
		return fmt.Sprintf("%.1f jam", duration.Hours())
	}
}
