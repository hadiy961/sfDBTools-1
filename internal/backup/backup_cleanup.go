// File : internal/backup/backup_cleanup.go
// Deskripsi : Fungsi terpadu untuk cleanup backup berdasarkan retention policy.
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-14
// Last Modified : 2024-10-15 (Simplified)

package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
)

const (
	// timeFormat mendefinisikan format timestamp standar untuk logging.
	timeFormat = "2006-01-02 15:04:05"
)

var (
	// backupExtensions mendefinisikan ekstensi file yang dianggap sebagai file backup.
	backupExtensions = []string{".sql", ".gz", ".zst", ".lz4", ".enc"}
)

// CleanupOldBackups menjalankan proses penghapusan semua backup lama di direktori.
func (s *Service) CleanupOldBackups() error {
	return s.cleanupCore(false, "") // dryRun=false, tanpa pattern
}

// CleanupDryRun menampilkan preview semua backup lama yang akan dihapus.
func (s *Service) CleanupDryRun() error {
	return s.cleanupCore(true, "") // dryRun=true, tanpa pattern
}

// CleanupByPattern menjalankan proses penghapusan backup lama yang cocok dengan pattern.
func (s *Service) CleanupByPattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("pattern tidak boleh kosong untuk CleanupByPattern")
	}
	return s.cleanupCore(false, pattern) // dryRun=false, dengan pattern
}

// cleanupCore adalah fungsi inti terpadu untuk semua logika pembersihan.
func (s *Service) cleanupCore(dryRun bool, pattern string) error {
	if !s.BackupOptions.Cleanup.Enabled {
		s.Logger.Info("Cleanup tidak diaktifkan, melewati proses.")
		return nil
	}

	// Tentukan mode operasi untuk logging
	mode := "Menjalankan"
	if dryRun {
		mode = "Menjalankan DRY-RUN"
	}
	if pattern != "" {
		s.Logger.Infof("%s proses cleanup untuk pattern: %s", mode, pattern)
	} else {
		s.Logger.Infof("%s proses cleanup backup...", mode)
	}

	retentionDays := s.BackupOptions.Cleanup.RetentionDays
	if retentionDays <= 0 {
		s.Logger.Info("Retention days tidak valid, melewati proses.")
		return nil
	}

	s.Logger.Infof("Cleanup policy: hapus file backup lebih dari %d hari", retentionDays)
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	s.Logger.Infof("Cutoff time: %s", cutoffTime.Format(timeFormat))

	// Pindai file berdasarkan mode (dengan atau tanpa pattern)
	filesToDelete, err := s.scanFiles(s.BackupOptions.OutputDirectory, cutoffTime, pattern)
	if err != nil {
		return fmt.Errorf("gagal memindai file backup: %w", err)
	}

	if len(filesToDelete) == 0 {
		s.Logger.Info("Tidak ada file backup lama yang perlu dihapus.")
		return nil
	}

	if dryRun {
		s.logDryRunSummary(filesToDelete)
	} else {
		s.performDeletion(filesToDelete)
	}

	return nil
}

// scanFiles memilih metode pemindaian file (menyeluruh atau berdasarkan pola).
func (s *Service) scanFiles(baseDir string, cutoff time.Time, pattern string) ([]BackupFileInfo, error) {
	// Jika tidak ada pattern, kita buat pattern default untuk mencari semua file secara rekursif.
	// Tanda '**/*' berarti "semua file di semua sub-direktori".
	if pattern == "" {
		pattern = "**/*"
	}

	// Satu panggilan untuk menemukan semua file yang cocok, di mana pun lokasinya!
	paths, err := doublestar.Glob(os.DirFS(baseDir), pattern)
	if err != nil {
		return nil, fmt.Errorf("gagal memproses pattern glob %s: %w", pattern, err)
	}

	var filesToDelete []BackupFileInfo
	for _, path := range paths {
		// Karena Glob mengembalikan path relatif, kita gabungkan lagi dengan baseDir
		fullPath := filepath.Join(baseDir, path)

		info, err := os.Stat(fullPath)
		if err != nil {
			s.Logger.Errorf("Gagal mendapatkan info file %s: %v", fullPath, err)
			continue
		}

		// Lewati direktori dan file yang tidak sesuai kriteria
		if info.IsDir() || (pattern == "**/*" && !s.isBackupFile(fullPath)) {
			continue
		}

		if info.ModTime().Before(cutoff) {
			filesToDelete = append(filesToDelete, BackupFileInfo{
				Path:    fullPath,
				ModTime: info.ModTime(),
				Size:    info.Size(),
			})
		}
	}

	sort.Slice(filesToDelete, func(i, j int) bool {
		return filesToDelete[i].ModTime.Before(filesToDelete[j].ModTime)
	})

	return filesToDelete, nil
}

// performDeletion menghapus file-file yang ada dalam daftar.
func (s *Service) performDeletion(files []BackupFileInfo) {
	s.Logger.Infof("Ditemukan %d file backup lama yang akan dihapus.", len(files))

	var deletedCount int
	var totalFreedSize int64

	for _, file := range files {
		if err := os.Remove(file.Path); err != nil {
			s.Logger.Errorf("Gagal menghapus file %s: %v", file.Path, err)
			continue
		}
		deletedCount++
		totalFreedSize += file.Size
		s.Logger.Infof("Dihapus: %s (size: %s)", file.Path, s.formatFileSize(file.Size))
	}

	s.Logger.Infof("Cleanup selesai: %d file dihapus, total %s ruang dibebaskan.",
		deletedCount, s.formatFileSize(totalFreedSize))
}

// logDryRunSummary mencatat ringkasan file yang akan dihapus dalam mode dry-run.
func (s *Service) logDryRunSummary(files []BackupFileInfo) {
	s.Logger.Infof("DRY-RUN: Ditemukan %d file backup yang AKAN dihapus:", len(files))

	var totalSize int64
	for i, file := range files {
		totalSize += file.Size
		s.Logger.Infof("  [%d] %s (modified: %s, size: %s)",
			i+1,
			file.Path,
			file.ModTime.Format(timeFormat),
			s.formatFileSize(file.Size))
	}

	s.Logger.Infof("DRY-RUN: Total %d file dengan ukuran %s akan dibebaskan.",
		len(files), s.formatFileSize(totalSize))
	s.Logger.Info("DRY-RUN: Untuk menjalankan cleanup sebenarnya, jalankan tanpa flag --dry-run.")
}

// isBackupFile memeriksa apakah sebuah file dianggap sebagai file backup berdasarkan ekstensinya.
func (s *Service) isBackupFile(filename string) bool {
	lowerFilename := strings.ToLower(filename)
	for _, ext := range backupExtensions {
		if strings.HasSuffix(lowerFilename, ext) {
			return true
		}
	}
	return false
}

// formatFileSize mengubah ukuran file dalam bytes menjadi format yang mudah dibaca.
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
