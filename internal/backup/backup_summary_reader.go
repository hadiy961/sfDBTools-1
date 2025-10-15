// File : internal/backup/backup_summary_reader.go
// Deskripsi : Fungsi untuk membaca dan menampilkan summary backup dari file JSON
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025

package backup

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sfDBTools/pkg/ui"
	"sort"
	"strings"
	"time"
)

// ReadSummaryFromJSON membaca summary dari file JSON
func (s *Service) ReadSummaryFromJSON(jsonPath string) (*BackupSummary, error) {
	// Baca file JSON
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file JSON: %w", err)
	}

	// Parse JSON
	var summary BackupSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil, fmt.Errorf("gagal parse JSON: %w", err)
	}

	return &summary, nil
}

// ListAvailableSummaries menampilkan daftar summary yang tersedia
func (s *Service) ListAvailableSummaries() error {
	// Gunakan base directory dari config untuk summary
	baseDirectory := s.Config.Backup.Output.BaseDirectory
	summaryDir := filepath.Join(baseDirectory, "summaries")

	s.Logger.Debugf("BaseDirectory: %s", baseDirectory)
	s.Logger.Debugf("Mencari summary di direktori: %s", summaryDir)

	// Periksa apakah direktori ada
	if _, err := os.Stat(summaryDir); os.IsNotExist(err) {
		fmt.Println("❌ Tidak ada summary backup yang tersedia.")
		fmt.Printf("Direktori summary tidak ditemukan: %s\n", summaryDir)
		return nil
	}

	// Scan file JSON di direktori summary
	var summaryFiles []SummaryFileEntry
	err := filepath.WalkDir(summaryDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
			return nil
		}

		// Baca summary untuk mendapatkan info dasar
		summary, readErr := s.ReadSummaryFromJSON(path)
		if readErr != nil {
			s.Logger.Warnf("Gagal membaca summary %s: %v", path, readErr)
			return nil
		}

		// Dapatkan informasi file
		fileInfo, statErr := d.Info()
		if statErr != nil {
			s.Logger.Warnf("Gagal mendapatkan info file %s: %v", path, statErr)
			return nil
		}

		summaryFiles = append(summaryFiles, SummaryFileEntry{
			FileName:      d.Name(),
			FilePath:      path,
			BackupID:      summary.BackupID,
			Status:        summary.Status,
			BackupMode:    summary.BackupMode,
			Timestamp:     summary.Timestamp,
			Duration:      summary.Duration,
			DatabaseCount: summary.DatabaseStats.SuccessfulBackups,
			FailedCount:   summary.DatabaseStats.FailedBackups,
			TotalSize:     summary.OutputInfo.TotalSizeHuman,
			CreatedAt:     fileInfo.ModTime(),
		})

		return nil
	})

	if err != nil {
		return fmt.Errorf("gagal scan direktori summary: %w", err)
	}

	if len(summaryFiles) == 0 {
		fmt.Println("❌ Tidak ada file summary backup yang valid ditemukan.")
		return nil
	}

	// Sort berdasarkan timestamp terbaru dulu
	sort.Slice(summaryFiles, func(i, j int) bool {
		return summaryFiles[i].Timestamp.After(summaryFiles[j].Timestamp)
	})

	// Tampilkan dalam format table
	ui.PrintHeader("DAFTAR SUMMARY BACKUP")

	fmt.Printf("📁 Direktori Summary: %s\n\n", summaryDir)

	headers := []string{"Backup ID", "Status", "Mode", "Tanggal", "Durasi", "DB Success/Failed", "Total Size"}
	var rows [][]string

	for _, entry := range summaryFiles {
		statusIcon := ui.GetStatusIcon(entry.Status)
		dbInfo := fmt.Sprintf("%d/%d", entry.DatabaseCount, entry.FailedCount)

		rows = append(rows, []string{
			entry.BackupID,
			statusIcon + " " + entry.Status,
			entry.BackupMode,
			entry.Timestamp.Format("2006-01-02 15:04"),
			entry.Duration,
			dbInfo,
			entry.TotalSize,
		})
	}

	ui.FormatTable(headers, rows)

	fmt.Printf("\n💡 Untuk melihat detail summary: gunakan --backup-id=<backup_id>\n")
	fmt.Printf("💡 Atau gunakan --latest untuk melihat summary terbaru\n\n")

	return nil
}

// ShowSummaryByID menampilkan summary berdasarkan backup ID
func (s *Service) ShowSummaryByID(backupID string) error {
	// Gunakan base directory dari config untuk summary
	baseDirectory := s.Config.Backup.Output.BaseDirectory
	summaryDir := filepath.Join(baseDirectory, "summaries")
	summaryFile := filepath.Join(summaryDir, backupID+".json")

	// Periksa apakah file ada
	if _, err := os.Stat(summaryFile); os.IsNotExist(err) {
		return fmt.Errorf("summary dengan ID '%s' tidak ditemukan: %s", backupID, summaryFile)
	}

	// Baca dan tampilkan summary
	summary, err := s.ReadSummaryFromJSON(summaryFile)
	if err != nil {
		return fmt.Errorf("gagal membaca summary: %w", err)
	}

	s.DisplaySummaryTable(summary)
	return nil
}

// ShowLatestSummary menampilkan summary backup terbaru
func (s *Service) ShowLatestSummary() error {
	// Gunakan base directory dari config untuk summary
	baseDirectory := s.Config.Backup.Output.BaseDirectory
	summaryDir := filepath.Join(baseDirectory, "summaries")

	// Periksa apakah direktori ada
	if _, err := os.Stat(summaryDir); os.IsNotExist(err) {
		return fmt.Errorf("tidak ada summary backup yang tersedia: direktori %s tidak ditemukan", summaryDir)
	}

	// Cari file JSON terbaru
	var latestFile string
	var latestTime time.Time

	err := filepath.WalkDir(summaryDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".json") {
			return nil
		}

		fileInfo, statErr := d.Info()
		if statErr != nil {
			return nil
		}

		if fileInfo.ModTime().After(latestTime) {
			latestTime = fileInfo.ModTime()
			latestFile = path
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("gagal scan direktori summary: %w", err)
	}

	if latestFile == "" {
		return fmt.Errorf("tidak ada file summary yang ditemukan")
	}

	// Baca dan tampilkan summary terbaru
	summary, err := s.ReadSummaryFromJSON(latestFile)
	if err != nil {
		return fmt.Errorf("gagal membaca summary terbaru: %w", err)
	}

	fmt.Printf("📋 Menampilkan Summary Backup Terbaru: %s\n\n", summary.BackupID)
	s.DisplaySummaryTable(summary)
	return nil
}
