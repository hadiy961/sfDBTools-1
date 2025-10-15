// File : internal/backup/backup_summary.go
// Deskripsi : Fungsi untuk membuat summary backup dalam format JSON dan table
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025

package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/pkg/ui"
	"time"
)

// BackupSummary adalah struktur untuk menyimpan summary backup
type BackupSummary struct {
	// Informasi umum backup
	BackupID   string    `json:"backup_id"`
	Timestamp  time.Time `json:"timestamp"`
	BackupMode string    `json:"backup_mode"` // "separate" atau "combined"
	Status     string    `json:"status"`      // "success", "partial", "failed"
	Duration   string    `json:"duration"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`

	// Informasi database
	DatabaseStats DatabaseSummaryStats `json:"database_stats"`

	// Informasi file output
	OutputInfo OutputSummaryInfo `json:"output_info"`

	// Konfigurasi backup
	BackupConfig BackupConfigSummary `json:"backup_config"`

	// Database yang berhasil dan gagal
	SuccessfulDatabases []DatabaseBackupInfo `json:"successful_databases"`
	FailedDatabases     []FailedDatabaseInfo `json:"failed_databases"`

	// Informasi error (jika ada)
	Errors []string `json:"errors,omitempty"`
}

// DatabaseSummaryStats berisi statistik database
type DatabaseSummaryStats struct {
	TotalDatabases    int `json:"total_databases"`
	SuccessfulBackups int `json:"successful_backups"`
	FailedBackups     int `json:"failed_backups"`
	ExcludedDatabases int `json:"excluded_databases"`
	SystemDatabases   int `json:"system_databases"`
	FilteredDatabases int `json:"filtered_databases"`
}

// OutputSummaryInfo berisi informasi file output
type OutputSummaryInfo struct {
	OutputDirectory string            `json:"output_directory"`
	TotalFiles      int               `json:"total_files"`
	TotalSize       int64             `json:"total_size_bytes"`
	TotalSizeHuman  string            `json:"total_size_human"`
	Files           []SummaryFileInfo `json:"files"`
}

// BackupConfigSummary berisi ringkasan konfigurasi backup
type BackupConfigSummary struct {
	CompressionEnabled bool   `json:"compression_enabled"`
	CompressionType    string `json:"compression_type,omitempty"`
	CompressionLevel   string `json:"compression_level,omitempty"`
	EncryptionEnabled  bool   `json:"encryption_enabled"`
	DBListFile         string `json:"db_list_file,omitempty"`
	CleanupEnabled     bool   `json:"cleanup_enabled"`
	RetentionDays      int    `json:"retention_days,omitempty"`
}

// DatabaseBackupInfo berisi informasi database yang berhasil dibackup
type DatabaseBackupInfo struct {
	DatabaseName  string `json:"database_name"`
	OutputFile    string `json:"output_file"`
	FileSize      int64  `json:"file_size_bytes"`
	FileSizeHuman string `json:"file_size_human"`
	Duration      string `json:"duration"`
}

// FailedDatabaseInfo berisi informasi database yang gagal dibackup
type FailedDatabaseInfo struct {
	DatabaseName string `json:"database_name"`
	Error        string `json:"error"`
}

// SummaryFileInfo berisi informasi file backup untuk summary (berbeda dari BackupFileInfo di cleanup)
type SummaryFileInfo struct {
	FileName     string    `json:"file_name"`
	FilePath     string    `json:"file_path"`
	Size         int64     `json:"size_bytes"`
	SizeHuman    string    `json:"size_human"`
	DatabaseName string    `json:"database_name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateBackupSummary membuat summary backup
func (s *Service) CreateBackupSummary(backupMode string, dbFiltered []string, successfulDBs []DatabaseBackupInfo, failedDBs []FailedDatabaseInfo, startTime time.Time, errors []string) *BackupSummary {
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// Generate backup ID
	backupID := fmt.Sprintf("backup_%s", startTime.Format("20060102_150405"))

	// Tentukan status backup
	status := "success"
	if len(failedDBs) > 0 {
		if len(successfulDBs) > 0 {
			status = "partial"
		} else {
			status = "failed"
		}
	}

	// Hitung statistik database
	dbStats := DatabaseSummaryStats{
		TotalDatabases:    len(dbFiltered),
		SuccessfulBackups: len(successfulDBs),
		FailedBackups:     len(failedDBs),
	}

	// Tambahkan informasi filter jika ada
	if s.FilterInfo != nil {
		dbStats.ExcludedDatabases = s.FilterInfo.ExcludedDatabases
		dbStats.SystemDatabases = s.FilterInfo.SystemDatabases
		dbStats.FilteredDatabases = s.FilterInfo.IncludedDatabases
	}

	// Kumpulkan informasi file output
	outputInfo := s.collectOutputInfo(successfulDBs)

	// Buat ringkasan konfigurasi backup
	backupConfig := BackupConfigSummary{
		CompressionEnabled: s.BackupOptions.Compression.Enabled,
		EncryptionEnabled:  s.BackupOptions.Encryption.Enabled,
		CleanupEnabled:     s.BackupOptions.Cleanup.Enabled,
	}

	if s.BackupOptions.Compression.Enabled {
		backupConfig.CompressionType = s.BackupOptions.Compression.Type
		backupConfig.CompressionLevel = s.BackupOptions.Compression.Level
	}

	if s.BackupOptions.DBList != "" {
		backupConfig.DBListFile = s.BackupOptions.DBList
	}

	if s.BackupOptions.Cleanup.Enabled {
		backupConfig.RetentionDays = s.BackupOptions.Cleanup.RetentionDays
	}

	return &BackupSummary{
		BackupID:            backupID,
		Timestamp:           startTime,
		BackupMode:          backupMode,
		Status:              status,
		Duration:            formatDuration(duration),
		StartTime:           startTime,
		EndTime:             endTime,
		DatabaseStats:       dbStats,
		OutputInfo:          outputInfo,
		BackupConfig:        backupConfig,
		SuccessfulDatabases: successfulDBs,
		FailedDatabases:     failedDBs,
		Errors:              errors,
	}
}

// collectOutputInfo mengumpulkan informasi file output
func (s *Service) collectOutputInfo(successfulDBs []DatabaseBackupInfo) OutputSummaryInfo {
	var files []SummaryFileInfo
	var totalSize int64

	outputDir := s.BackupOptions.OutputDirectory

	for _, db := range successfulDBs {
		fileInfo := SummaryFileInfo{
			FileName:     filepath.Base(db.OutputFile),
			FilePath:     db.OutputFile,
			Size:         db.FileSize,
			SizeHuman:    db.FileSizeHuman,
			DatabaseName: db.DatabaseName,
			CreatedAt:    time.Now(), // Bisa diambil dari file stat jika perlu akurat
		}
		files = append(files, fileInfo)
		totalSize += db.FileSize
	}

	return OutputSummaryInfo{
		OutputDirectory: outputDir,
		TotalFiles:      len(files),
		TotalSize:       totalSize,
		TotalSizeHuman:  formatFileSize(totalSize),
		Files:           files,
	}
}

// SaveSummaryToJSON menyimpan summary ke file JSON
func (s *Service) SaveSummaryToJSON(summary *BackupSummary) error {
	// Gunakan base directory dari config untuk summary
	baseDirectory := s.Config.Backup.Output.BaseDirectory
	summaryDir := filepath.Join(baseDirectory, "summaries")
	if err := os.MkdirAll(summaryDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori summary: %w", err)
	}

	// Generate nama file summary
	summaryFileName := fmt.Sprintf("%s.json", summary.BackupID)
	summaryPath := filepath.Join(summaryDir, summaryFileName)

	// Convert ke JSON dengan indentasi
	jsonData, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("gagal marshal summary ke JSON: %w", err)
	}

	// Tulis ke file
	if err := os.WriteFile(summaryPath, jsonData, 0644); err != nil {
		return fmt.Errorf("gagal menulis file summary: %w", err)
	}

	s.Logger.Infof("Summary backup disimpan ke: %s", summaryPath)
	return nil
}

// DisplaySummaryTable menampilkan summary dalam format table
func (s *Service) DisplaySummaryTable(summary *BackupSummary) {
	ui.PrintHeader("BACKUP SUMMARY")

	// Tabel informasi umum
	fmt.Println("📊 Informasi Umum:")
	generalHeaders := []string{"Property", "Value"}
	generalRows := [][]string{
		{"Backup ID", summary.BackupID},
		{"Status", getStatusIcon(summary.Status) + " " + summary.Status},
		{"Mode", summary.BackupMode},
		{"Waktu Mulai", summary.StartTime.Format("2006-01-02 15:04:05")},
		{"Waktu Selesai", summary.EndTime.Format("2006-01-02 15:04:05")},
		{"Durasi", summary.Duration},
	}
	ui.FormatTable(generalHeaders, generalRows)
	fmt.Println()

	// Tabel statistik database
	fmt.Println("🗄️  Statistik Database:")
	dbStatsHeaders := []string{"Metric", "Count"}
	dbStatsRows := [][]string{
		{"Total Database", fmt.Sprintf("%d", summary.DatabaseStats.TotalDatabases)},
		{"Berhasil", fmt.Sprintf("✅ %d", summary.DatabaseStats.SuccessfulBackups)},
		{"Gagal", fmt.Sprintf("❌ %d", summary.DatabaseStats.FailedBackups)},
	}
	if summary.DatabaseStats.ExcludedDatabases > 0 {
		dbStatsRows = append(dbStatsRows, []string{"Dikecualikan", fmt.Sprintf("⚠️ %d", summary.DatabaseStats.ExcludedDatabases)})
	}
	ui.FormatTable(dbStatsHeaders, dbStatsRows)
	fmt.Println()

	// Tabel informasi file output
	fmt.Println("📁 Output Files:")
	outputHeaders := []string{"Property", "Value"}
	outputRows := [][]string{
		{"Direktori Output", summary.OutputInfo.OutputDirectory},
		{"Total File", fmt.Sprintf("%d", summary.OutputInfo.TotalFiles)},
		{"Total Ukuran", summary.OutputInfo.TotalSizeHuman},
	}
	ui.FormatTable(outputHeaders, outputRows)
	fmt.Println()

	// Tabel konfigurasi backup
	fmt.Println("⚙️  Konfigurasi Backup:")
	configHeaders := []string{"Fitur", "Status", "Detail"}
	var configRows [][]string

	// Kompresi
	compressionStatus := "❌ Disabled"
	compressionDetail := "-"
	if summary.BackupConfig.CompressionEnabled {
		compressionStatus = "✅ Enabled"
		compressionDetail = fmt.Sprintf("%s (level: %s)", summary.BackupConfig.CompressionType, summary.BackupConfig.CompressionLevel)
	}
	configRows = append(configRows, []string{"Kompresi", compressionStatus, compressionDetail})

	// Enkripsi
	encryptionStatus := "❌ Disabled"
	encryptionDetail := "-"
	if summary.BackupConfig.EncryptionEnabled {
		encryptionStatus = "✅ Enabled"
		encryptionDetail = "AES-256-GCM"
	}
	configRows = append(configRows, []string{"Enkripsi", encryptionStatus, encryptionDetail})

	// Cleanup
	cleanupStatus := "❌ Disabled"
	cleanupDetail := "-"
	if summary.BackupConfig.CleanupEnabled {
		cleanupStatus = "✅ Enabled"
		cleanupDetail = fmt.Sprintf("%d hari", summary.BackupConfig.RetentionDays)
	}
	configRows = append(configRows, []string{"Auto Cleanup", cleanupStatus, cleanupDetail})

	ui.FormatTable(configHeaders, configRows)
	fmt.Println()

	// Tabel database yang berhasil (jika ada)
	if len(summary.SuccessfulDatabases) > 0 {
		fmt.Println("✅ Database Berhasil:")
		successHeaders := []string{"Database", "File Output", "Ukuran", "Durasi"}
		var successRows [][]string

		for _, db := range summary.SuccessfulDatabases {
			successRows = append(successRows, []string{
				db.DatabaseName,
				filepath.Base(db.OutputFile),
				db.FileSizeHuman,
				db.Duration,
			})
		}
		ui.FormatTable(successHeaders, successRows)
		fmt.Println()
	}

	// Tabel database yang gagal (jika ada)
	if len(summary.FailedDatabases) > 0 {
		fmt.Println("❌ Database Gagal:")
		failedHeaders := []string{"Database", "Error"}
		var failedRows [][]string

		for _, db := range summary.FailedDatabases {
			failedRows = append(failedRows, []string{db.DatabaseName, db.Error})
		}
		ui.FormatTable(failedHeaders, failedRows)
		fmt.Println()
	}

	// Tampilkan error umum (jika ada)
	if len(summary.Errors) > 0 {
		fmt.Println("⚠️  Error Umum:")
		for _, err := range summary.Errors {
			fmt.Printf("   • %s\n", err)
		}
		fmt.Println()
	}
}

// Helper functions

// getStatusIcon mengembalikan icon untuk status
func getStatusIcon(status string) string {
	switch status {
	case "success":
		return "✅"
	case "partial":
		return "⚠️"
	case "failed":
		return "❌"
	default:
		return "❓"
	}
}

// formatDuration memformat durasi menjadi string yang mudah dibaca
func formatDuration(duration time.Duration) string {
	if duration < time.Minute {
		return fmt.Sprintf("%.1f detik", duration.Seconds())
	} else if duration < time.Hour {
		return fmt.Sprintf("%.1f menit", duration.Minutes())
	} else {
		return fmt.Sprintf("%.1f jam", duration.Hours())
	}
}

// formatFileSize memformat ukuran file menjadi string yang mudah dibaca
func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatFileSize adalah wrapper untuk formatFileSize agar bisa dipanggil dari luar
func (s *Service) FormatFileSize(bytes int64) string {
	return formatFileSize(bytes)
}

// FormatDuration adalah wrapper untuk formatDuration agar bisa dipanggil dari luar
func (s *Service) FormatDuration(duration time.Duration) string {
	return formatDuration(duration)
}
