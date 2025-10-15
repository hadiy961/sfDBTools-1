// File : internal/backup/backup_summary_reader.go
// Deskripsi : Fungsi untuk membaca dan menampilkan summary backup dari file JSON
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025 (Refactored for clarity, efficiency, and SRP)

package backup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/pkg/ui"
	"sort"
	"strings"
)

// ListAvailableSummaries menampilkan daftar summary yang tersedia.
func (s *Service) ListAvailableSummaries() error {
	summaryFiles, err := s.findAllSummaries()
	if err != nil {
		// Jika error karena direktori tidak ada, berikan pesan yang lebih ramah.
		if os.IsNotExist(err) {
			ui.PrintError("Tidak ada summary backup yang tersedia.")
			ui.PrintInfo(fmt.Sprintf("Direktori summary tidak ditemukan di: %s", s.getSummaryDir()))
			return nil
		}
		return fmt.Errorf("gagal mendapatkan daftar summary: %w", err)
	}

	if len(summaryFiles) == 0 {
		ui.PrintError("Tidak ada file summary backup yang valid ditemukan.")
		return nil
	}

	// Sort berdasarkan timestamp terbaru dulu
	sort.Slice(summaryFiles, func(i, j int) bool {
		return summaryFiles[i].Timestamp.After(summaryFiles[j].Timestamp)
	})

	s.displaySummaryList(summaryFiles)
	return nil
}

// ShowSummaryByID menampilkan summary berdasarkan backup ID.
func (s *Service) ShowSummaryByID(backupID string) error {
	summaryFile := filepath.Join(s.getSummaryDir(), backupID+".json")

	// Periksa apakah file ada sebelum membaca.
	if _, err := os.Stat(summaryFile); os.IsNotExist(err) {
		return fmt.Errorf("summary dengan ID '%s' tidak ditemukan di %s", backupID, summaryFile)
	}

	summary, err := s.readSummaryFromJSON(summaryFile)
	if err != nil {
		return fmt.Errorf("gagal membaca summary: %w", err)
	}

	s.DisplaySummaryTable(summary)
	return nil
}

// ShowLatestSummary menampilkan summary backup terbaru.
func (s *Service) ShowLatestSummary() error {
	summaries, err := s.findAllSummaries()
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("tidak ada summary backup yang tersedia: direktori %s tidak ditemukan", s.getSummaryDir())
		}
		return fmt.Errorf("gagal mencari summary: %w", err)
	}

	if len(summaries) == 0 {
		return fmt.Errorf("tidak ada file summary yang ditemukan")
	}

	// Sort untuk memastikan yang pertama adalah yang terbaru.
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Timestamp.After(summaries[j].Timestamp)
	})

	latestSummaryPath := summaries[0].FilePath
	s.Logger.Infof("Menemukan summary terbaru: %s", latestSummaryPath)

	summary, err := s.readSummaryFromJSON(latestSummaryPath)
	if err != nil {
		return fmt.Errorf("gagal membaca summary terbaru: %w", err)
	}

	ui.PrintInfo(fmt.Sprintf("Menampilkan Summary Backup Terbaru: %s", summary.BackupID))
	s.DisplaySummaryTable(summary)
	return nil
}

// findAllSummaries adalah fungsi helper terpusat untuk memindai dan membaca semua metadata summary.
// Ini menghilangkan duplikasi kode antara ListAvailableSummaries dan ShowLatestSummary.
func (s *Service) findAllSummaries() ([]SummaryFileEntry, error) {
	summaryDir := s.getSummaryDir()
	s.Logger.Debugf("Mencari summary di direktori: %s", summaryDir)

	entries, err := os.ReadDir(summaryDir)
	if err != nil {
		return nil, err // Kembalikan error asli (misal: os.IsNotExist)
	}

	var summaryFiles []SummaryFileEntry
	for _, entry := range entries {
		// Lewati direktori atau file yang bukan .json
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".json") {
			continue
		}

		filePath := filepath.Join(summaryDir, entry.Name())
		summary, readErr := s.readSummaryFromJSON(filePath)
		if readErr != nil {
			s.Logger.Warnf("Gagal memproses file summary %s: %v", entry.Name(), readErr)
			continue // Lanjut ke file berikutnya jika ada yang korup
		}

		fileInfo, statErr := entry.Info()
		if statErr != nil {
			s.Logger.Warnf("Gagal mendapatkan info file untuk %s: %v", entry.Name(), statErr)
			continue
		}

		summaryFiles = append(summaryFiles, SummaryFileEntry{
			FileName:      entry.Name(),
			FilePath:      filePath,
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
	}
	return summaryFiles, nil
}

// displaySummaryList bertanggung jawab HANYA untuk menampilkan daftar summary ke UI.
func (s *Service) displaySummaryList(summaries []SummaryFileEntry) {
	ui.PrintHeader("DAFTAR SUMMARY BACKUP")
	ui.PrintInfo(fmt.Sprintf("Direktori Summary: %s\n", s.getSummaryDir()))

	headers := []string{"Backup ID", "Status", "Mode", "Tanggal", "Durasi", "DB Success/Failed", "Total Size"}
	var rows [][]string

	for _, entry := range summaries {
		rows = append(rows, []string{
			entry.BackupID,
			ui.GetStatusIcon(entry.Status) + " " + entry.Status,
			entry.BackupMode,
			entry.Timestamp.Format("2006-01-02 15:04"),
			entry.Duration,
			fmt.Sprintf("%d/%d", entry.DatabaseCount, entry.FailedCount),
			entry.TotalSize,
		})
	}

	ui.FormatTable(headers, rows)

	fmt.Printf("\n💡 Untuk melihat detail summary: %s\n", ui.ColorText("sfdbtools summary --backup-id=<backup_id>", ui.ColorCyan))
	fmt.Printf("💡 Atau gunakan %s untuk melihat summary terbaru.\n\n", ui.ColorText("sfdbtools summary --latest", ui.ColorCyan))
}

// readSummaryFromJSON membaca dan mengurai file summary JSON.
// Nama diubah menjadi huruf kecil karena ini adalah helper internal.
func (s *Service) readSummaryFromJSON(jsonPath string) (*BackupSummary, error) {
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file %s: %w", jsonPath, err)
	}

	var summary BackupSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil, fmt.Errorf("gagal parse JSON dari %s: %w", jsonPath, err)
	}

	return &summary, nil
}

// getSummaryDir adalah helper untuk mendapatkan path direktori summary secara konsisten.
func (s *Service) getSummaryDir() string {
	return filepath.Join(s.Config.Backup.Output.BaseDirectory, "summaries")
}
