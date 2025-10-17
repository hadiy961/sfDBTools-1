// File : internal/backup/backup_utils.go
// Deskripsi : Utility functions untuk modul backup (helper dan display functions)
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-08
// Last Modified : 16 Oktober 2025

package backup

import (
	"context"
	"fmt"
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

func (s *Service) getTargetDBConfig() structs.ServerDBConnection {
	// Jika konfigurasi target DB sudah ada di ScanOptions, gunakan itu.
	// Jika tidak, fallback ke environment variables.

	return structs.ServerDBConnection{
		Host:     database.GetEnvOrDefault("SFDB_DB_HOST", "localhost"),
		Port:     database.GetEnvOrDefaultInt("SFDB_DB_PORT", 3306),
		User:     database.GetEnvOrDefault("SFDB_DB_USER", "root"),
		Password: database.GetEnvOrDefault("SFDB_DB_PASSWORD", ""),
		Database: database.GetEnvOrDefault("SFDB_DB_NAME", ""),
	}
}

func (s *Service) ConnectToTargetDB(ctx context.Context) (*database.Client, error) {
	targetConn := s.getTargetDBConfig()

	client, err := database.InitializeDatabase(targetConn)
	if err != nil {
		return nil, fmt.Errorf("gagal koneksi ke target database: %w", err)
	}

	// Verify koneksi dengan ping
	if err := client.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("gagal verifikasi koneksi: %w", err)
	}

	ui.PrintSuccess(fmt.Sprintf("Koneksi ke target database berhasil: %s@%s:%d/%s",
		targetConn.User, targetConn.Host, targetConn.Port, targetConn.Database))

	return client, nil
}
