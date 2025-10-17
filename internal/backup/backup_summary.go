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
	"path/filepath" // <-- Menggunakan package UI Anda
	"sfDBTools/pkg/ui"
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
		Duration:            ui.FormatDuration(endTime.Sub(startTime)),
		StartTime:           startTime,
		EndTime:             endTime,
		DatabaseStats:       s.buildDatabaseStats(dbFiltered, successfulDBs, failedDBs),
		OutputInfo:          s.collectOutputInfo(successfulDBs),
		BackupConfig:        s.buildBackupConfigSummary(),
		SuccessfulDatabases: successfulDBs,
		FailedDatabases:     failedDBs,
		DatabaseDetails:     s.DatabaseDetail,
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
