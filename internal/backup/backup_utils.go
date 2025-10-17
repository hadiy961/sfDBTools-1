// File : internal/backup/backup_utils.go
// Deskripsi : Utility functions untuk modul backup (helper dan display functions)
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-08
// Last Modified : 16 Oktober 2025

package backup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sfDBTools/internal/dbscan"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
	"strconv"
	"strings"
)

// ===== HELPER FUNCTIONS =====

// sanitizeArgsForLogging menyembunyikan password dalam argumen untuk logging
func (s *Service) sanitizeArgsForLogging(args []string) []string {
	sanitized := make([]string, len(args))
	copy(sanitized, args)

	for i, arg := range sanitized {
		if strings.HasPrefix(arg, "--password=") {
			sanitized[i] = "--password=***"
		}
	}

	return sanitized
}

// ===== DISPLAY FUNCTIONS =====

// DisplayBackupAllOptions menampilkan opsi backup yang digunakan
func (s *Service) DisplayBackupAllOptions() {
	ui.PrintSubHeader("Opsi Backup Database All")
	// menampilkan seluruh opsi backup yang digunakan dalam bentuk table
	headers := []string{"Option", "Value"}
	port := s.DBConfigInfo.ServerDBConnection.Port
	data := [][]string{
		{"Configuration Path", s.DBConfigInfo.FilePath},
		{"Host", s.DBConfigInfo.ServerDBConnection.Host},
		{"Port", strconv.Itoa(port)},
		{"Username", s.DBConfigInfo.ServerDBConnection.User},
		{"Backup Directory", s.BackupOptions.OutputDirectory},
		{"Compression", s.BackupOptions.Compression.Type},
		{"Compression Level", (s.BackupOptions.Compression.Level)},
		{"Encryption Enabled", strconv.FormatBool(s.BackupOptions.Encryption.Enabled)},
		{"Cleanup Enabled", strconv.FormatBool(s.BackupOptions.Cleanup.Enabled)},
		{"Cleanup Schedule", s.BackupOptions.Cleanup.Scheduled},
		{"Retention Days", strconv.Itoa(s.BackupOptions.Cleanup.RetentionDays)},
		{"Exclude Databases", ui.FormatStringSlice(s.BackupOptions.Exclude.Databases)},
		{"Exclude Users", strconv.FormatBool(s.BackupOptions.Exclude.Users)},
		{"Exclude System Databases", strconv.FormatBool(s.BackupOptions.Exclude.SystemsDB)},
		{"Exclude Data", strconv.FormatBool(s.BackupOptions.Exclude.Data)},
		{"Use DBList File", strconv.FormatBool(s.BackupOptions.UseDBList)},
		{"Database List File", s.BackupOptions.DBList},
		{"Verification Disk Check", strconv.FormatBool(s.BackupOptions.DiskCheck)},
		{"Capture GTID", strconv.FormatBool(s.BackupAll != nil && s.BackupAll.CaptureGtid)},
		{"Create Backup Info File", strconv.FormatBool(s.BackupInfo.Enabled)},
	}
	ui.FormatTable(headers, data)
}

// addFileExtensions menambahkan ekstensi kompresi dan enkripsi ke nama file
func (s *Service) addFileExtensions(filename string, config BackupConfig) string {
	// Tambahkan ekstensi kompresi jika diperlukan
	if config.CompressionRequired {
		extension := compress.GetFileExtension(compress.CompressionType(config.CompressionType))
		filename += extension
	}

	// Tambahkan ekstensi enkripsi jika diperlukan
	if config.EncryptionEnabled {
		filename += ".enc"
	}

	return filename
}

// getShortage menghitung kekurangan ruang disk jika tidak cukup.
func (s *Service) getShortage(r *structs.DiskSpaceCheckResult) uint64 {
	if r.HasEnoughSpace || r.AvailableDiskSpace >= r.RequiredWithMargin {
		return 0
	}
	return r.RequiredWithMargin - r.AvailableDiskSpace
}

// isFatalMysqldumpError menentukan apakah error dari mysqldump adalah fatal atau hanya warning
// Fatal errors: koneksi gagal, permission denied, database tidak ada, dll
// Non-fatal: view errors, trigger errors (data masih bisa di-backup)
func (s *Service) isFatalMysqldumpError(err error, stderrOutput string) bool {
	if err == nil {
		return false
	}

	// Cek exit code jika tersedia
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitCode := exitErr.ExitCode()
		// mysqldump exit code 0 = success, 2 = warning (non-fatal)
		if exitCode == 2 {
			return false
		}
	}

	// Cek pattern error yang non-fatal (warnings yang tidak menghentikan dump)
	nonFatalPatterns := []string{
		"Couldn't read keys from table",
		"references invalid table(s) or column(s) or function(s)",
		"definer/invoker of view lack rights",
		"Warning:",
	}

	stderrLower := strings.ToLower(stderrOutput)
	for _, pattern := range nonFatalPatterns {
		if strings.Contains(stderrLower, strings.ToLower(pattern)) {
			// Jika hanya ada warning patterns dan bukan error fatal lainnya
			// Cek juga apakah ada error fatal
			if !strings.Contains(stderrLower, "access denied") &&
				!strings.Contains(stderrLower, "unknown database") &&
				!strings.Contains(stderrLower, "can't connect") &&
				!strings.Contains(stderrLower, "connection refused") {
				return false
			}
		}
	}

	// Default: anggap fatal
	return true
}

// saveErrorLog menyimpan stderr output ke file log
// Mengembalikan path file log yang dibuat
func (s *Service) saveErrorLog(outputDir, dbName, errorOutput string) string {
	if errorOutput == "" {
		return ""
	}

	// Generate nama file log
	logFileName := dbName + "_errors.log"
	logFilePath := filepath.Join(outputDir, logFileName)

	// Simpan ke file
	err := os.WriteFile(logFilePath, []byte(errorOutput), 0644)
	if err != nil {
		s.Logger.Errorf("Gagal menyimpan error log ke %s: %v", logFilePath, err)
		return ""
	}

	return logFilePath
}

// runDatabaseScanForBackup menjalankan database scan untuk mengumpulkan detail database yang hilang
// sebelum proses backup dimulai. Fungsi ini menggunakan dbscan service untuk melakukan scan.
func (s *Service) runDatabaseScanForBackup(ctx context.Context, dbNames []string) error {
	s.Logger.Info("Mempersiapkan database scan...")

	// Buat dbscan service dengan konfigurasi dari backup service
	dbscanSvc := dbscan.NewService(s.Logger, s.Config)

	// Setup scan options dari backup service context
	scanOptions := structs.ScanOptions{
		DBConfig:       *s.DBConfigInfo,
		ExcludeSystem:  false, // Tidak exclude system database karena kita scan specific databases
		IncludeList:    dbNames,
		SaveToDB:       true, // Simpan hasil scan ke database
		Background:     false,
		Mode:           "database",
		DisplayResults: true,
	}

	dbscanSvc.SetScanOptions(scanOptions)

	s.Logger.Infof("Melakukan scan untuk %d database...", len(dbNames))
	ui.PrintSubHeader("Database Scan - Mengumpulkan Detail Database")

	// Setup connections menggunakan fungsi internal dbscan
	sourceClient, err := database.InitializeDatabase(s.DBConfigInfo.ServerDBConnection)
	if err != nil {
		return fmt.Errorf("gagal inisialisasi koneksi source: %w", err)
	}
	defer sourceClient.Close()

	// Koneksi ke target database untuk menyimpan hasil
	targetClient, err := dbscanSvc.ConnectToTargetDB(ctx)
	if err != nil {
		return fmt.Errorf("gagal koneksi ke target database: %w", err)
	}
	defer targetClient.Close()

	// Jalankan scan
	result, err := dbscanSvc.ExecuteScan(ctx, sourceClient, targetClient, dbNames, false)
	if err != nil {
		return fmt.Errorf("gagal menjalankan scan: %w", err)
	}

	// Log hasil scan
	s.Logger.Infof("Scan selesai - Berhasil: %d, Gagal: %d", result.SuccessCount, result.FailedCount)

	if result.FailedCount > 0 {
		s.Logger.Warnf("Beberapa database gagal di-scan (%d database)", result.FailedCount)
		if len(result.Errors) > 0 {
			for i, errMsg := range result.Errors {
				s.Logger.Debugf("  %d. %s", i+1, errMsg)
			}
		}
	}

	ui.PrintSuccess(fmt.Sprintf("Database scan selesai - %d database berhasil di-scan", result.SuccessCount))

	return nil
}
