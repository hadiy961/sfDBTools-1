// File : internal/backup/backup_cleanup.go
// Deskripsi : Fungsi untuk cleanup backup berdasarkan retention policy
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-14
// Last Modified : 2024-10-14

package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// CleanupOldBackups membersihkan backup lama berdasarkan retention policy
func (s *Service) CleanupOldBackups() error {
	// Cek apakah cleanup diaktifkan
	if !s.BackupOptions.Cleanup.Enabled {
		s.Logger.Info("Cleanup tidak diaktifkan, melewati proses cleanup")
		return nil
	}

	s.Logger.Info("Memulai proses cleanup backup lama...")

	outputDir := s.BackupOptions.OutputDirectory
	retentionDays := s.BackupOptions.Cleanup.RetentionDays

	if retentionDays <= 0 {
		s.Logger.Info("Retention days tidak valid, melewati cleanup")
		return nil
	}

	s.Logger.Infof("Cleanup policy: hapus file backup lebih dari %d hari", retentionDays)

	// Tentukan cutoff time
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	s.Logger.Infof("Cutoff time: %s", cutoffTime.Format("2006-01-02 15:04:05"))

	// Scan direktori backup
	filesToDelete, err := s.scanBackupFiles(outputDir, cutoffTime)
	if err != nil {
		return fmt.Errorf("gagal scan file backup: %w", err)
	}

	if len(filesToDelete) == 0 {
		s.Logger.Info("Tidak ada file backup lama yang perlu dihapus")
		return nil
	}

	s.Logger.Infof("Ditemukan %d file backup lama yang akan dihapus", len(filesToDelete))

	// Hapus file-file lama
	deletedCount := 0
	var deletedSize int64

	for _, fileInfo := range filesToDelete {
		if err := s.deleteBackupFile(fileInfo); err != nil {
			s.Logger.Errorf("Gagal menghapus file %s: %v", fileInfo.Path, err)
			continue
		}
		deletedCount++
		deletedSize += fileInfo.Size
		s.Logger.Infof("Dihapus: %s (size: %s)", fileInfo.Path, s.formatFileSize(fileInfo.Size))
	}

	s.Logger.Infof("Cleanup selesai: %d file dihapus, total %s dibebaskan",
		deletedCount, s.formatFileSize(deletedSize))

	return nil
}

// BackupFileInfo menyimpan informasi file backup
type BackupFileInfo struct {
	Path    string
	ModTime time.Time
	Size    int64
}

// scanBackupFiles mencari file backup yang lebih lama dari cutoff time
func (s *Service) scanBackupFiles(baseDir string, cutoffTime time.Time) ([]BackupFileInfo, error) {
	var filesToDelete []BackupFileInfo

	err := filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			s.Logger.Errorf("Error accessing file %s: %v", path, err)
			return nil // Continue walking
		}

		// Skip direktori
		if info.IsDir() {
			return nil
		}

		// Cek apakah file adalah backup file (berdasarkan ekstensi)
		if !s.isBackupFile(path) {
			return nil
		}

		// Cek apakah file lebih lama dari cutoff time
		if info.ModTime().Before(cutoffTime) {
			filesToDelete = append(filesToDelete, BackupFileInfo{
				Path:    path,
				ModTime: info.ModTime(),
				Size:    info.Size(),
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort berdasarkan waktu modifikasi (terlama pertama)
	sort.Slice(filesToDelete, func(i, j int) bool {
		return filesToDelete[i].ModTime.Before(filesToDelete[j].ModTime)
	})

	return filesToDelete, nil
}

// isBackupFile memeriksa apakah file adalah backup file berdasarkan ekstensi
func (s *Service) isBackupFile(filename string) bool {
	// Ekstensi yang dianggap sebagai backup file
	backupExtensions := []string{".sql", ".gz", ".zst", ".lz4", ".enc"}

	filename = strings.ToLower(filename)

	for _, ext := range backupExtensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}

	// Cek kombinasi ekstensi (misal: .sql.gz.enc)
	if strings.Contains(filename, ".sql") {
		return true
	}

	return false
}

// deleteBackupFile menghapus file backup
func (s *Service) deleteBackupFile(fileInfo BackupFileInfo) error {
	if err := os.Remove(fileInfo.Path); err != nil {
		return err
	}
	return nil
}

// formatFileSize mengformat ukuran file ke string yang readable
func (s *Service) formatFileSize(bytes int64) string {
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

// CleanupByPattern membersihkan backup berdasarkan pattern nama file tertentu
func (s *Service) CleanupByPattern(pattern string) error {
	s.Logger.Infof("Cleanup backup dengan pattern: %s", pattern)

	outputDir := s.BackupOptions.OutputDirectory
	retentionDays := s.BackupOptions.Cleanup.RetentionDays

	if retentionDays <= 0 {
		return fmt.Errorf("retention days tidak valid: %d", retentionDays)
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	matches, err := filepath.Glob(filepath.Join(outputDir, pattern))
	if err != nil {
		return fmt.Errorf("gagal scan pattern %s: %w", pattern, err)
	}

	deletedCount := 0
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			s.Logger.Errorf("Gagal stat file %s: %v", match, err)
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(match); err != nil {
				s.Logger.Errorf("Gagal hapus file %s: %v", match, err)
				continue
			}
			deletedCount++
			s.Logger.Infof("Dihapus: %s", match)
		}
	}

	s.Logger.Infof("Cleanup pattern selesai: %d file dihapus", deletedCount)
	return nil
}

// CleanupDryRun menampilkan file yang akan dihapus tanpa menghapus
func (s *Service) CleanupDryRun() error {
	// Cek apakah cleanup diaktifkan
	if !s.BackupOptions.Cleanup.Enabled {
		s.Logger.Info("Cleanup tidak diaktifkan, melewati dry-run")
		return nil
	}

	s.Logger.Info("Menjalankan cleanup dry-run (preview mode)...")

	outputDir := s.BackupOptions.OutputDirectory
	retentionDays := s.BackupOptions.Cleanup.RetentionDays

	if retentionDays <= 0 {
		s.Logger.Info("Retention days tidak valid, melewati dry-run")
		return nil
	}

	s.Logger.Infof("Dry-run policy: file backup lebih dari %d hari", retentionDays)

	// Tentukan cutoff time
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	s.Logger.Infof("Cutoff time: %s", cutoffTime.Format("2006-01-02 15:04:05"))

	// Scan direktori backup
	filesToDelete, err := s.scanBackupFiles(outputDir, cutoffTime)
	if err != nil {
		return fmt.Errorf("gagal scan file backup: %w", err)
	}

	if len(filesToDelete) == 0 {
		s.Logger.Info("Tidak ada file backup lama yang akan dihapus")
		return nil
	}

	s.Logger.Infof("DRY-RUN: Ditemukan %d file backup yang AKAN dihapus:", len(filesToDelete))

	var totalSize int64
	for i, fileInfo := range filesToDelete {
		totalSize += fileInfo.Size
		s.Logger.Infof("  [%d] %s (modified: %s, size: %s)",
			i+1,
			fileInfo.Path,
			fileInfo.ModTime.Format("2006-01-02 15:04:05"),
			s.formatFileSize(fileInfo.Size))
	}

	s.Logger.Infof("DRY-RUN: Total %d file, %s akan dibebaskan",
		len(filesToDelete), s.formatFileSize(totalSize))
	s.Logger.Info("DRY-RUN: Untuk menjalankan cleanup sebenarnya, jalankan tanpa flag --dry-run")

	return nil
}
