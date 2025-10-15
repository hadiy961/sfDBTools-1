// File : internal/backup/backup_common.go
// Deskripsi : Common functions untuk mengurangi duplikasi kode backup
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-15
// Last Modified : 2024-10-15

package backup

import (
	"context"
	"log"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/fs"
	"sfDBTools/pkg/ui"
)

// BackupConfig menyimpan konfigurasi backup yang umum digunakan
type BackupConfig struct {
	BaseDumpArgs        string
	OutputDir           string
	CompressionType     string
	CompressionRequired bool
	EncryptionEnabled   bool
}

// PrepareBackupSession menangani setup awal yang sama untuk semua jenis backup
// Mengembalikan client, filtered databases, original max statements time, dan error
func (s *Service) PrepareBackupSession(ctx context.Context, headerTitle string, showOptions bool) (*database.Client, []string, float64, error) {
	ui.Headers(headerTitle)

	// Check flag configuration file
	if err := s.CheckAndSelectConfigFile(); err != nil {
		return nil, nil, 0, err
	}

	// Display options jika diminta
	if showOptions {
		s.DisplayBackupAllOptions()
	}

	// Membuat klien baru dengan semua konfigurasi di atas
	client, err := database.InitializeDatabase(s.DBConfigInfo.ServerDBConnection)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	// Get, Set dan cek max_statement_time untuk sesi ini
	originalMaxStatementsTime, err := s.AturMaxStatementsTime(ctx, client)
	if err != nil {
		s.Logger.Warn("Gagal mengatur max_statement_time: " + err.Error())
		client.Close()
		return nil, nil, 0, err
	}

	// Cek dan filter database yang akan di-backup
	dbFiltered, err := s.GetAndFilterDatabases(ctx, client)
	if err != nil {
		s.Logger.Error("Gagal mendapatkan dan memfilter database: " + err.Error())
		// Kembalikan nilai awal max_statement_time jika ada error
		s.KembalikanMaxStatementsTime(ctx, client, originalMaxStatementsTime)
		client.Close()
		return nil, nil, 0, err
	}

	return client, dbFiltered, originalMaxStatementsTime, nil
}

// SetupBackupExecution mempersiapkan konfigurasi backup yang umum
func (s *Service) SetupBackupExecution() (BackupConfig, error) {
	ui.PrintSubHeader("Persiapan Eksekusi Backup")

	// 1. Jalankan cleanup backup lama terlebih dahulu untuk membebaskan ruang disk
	s.Logger.Info("Menjalankan cleanup backup lama sebelum backup...")
	if err := s.CleanupOldBackups(); err != nil {
		s.Logger.Errorf("Cleanup backup lama gagal: %v", err)
		// Lanjutkan backup meskipun cleanup gagal
	}

	// 2. Pastikan output directory sudah ada dengan memanggil ValidateOutput
	if err := s.ValidateOutput(); err != nil {
		return BackupConfig{}, err
	}

	// 3. Setup konfigurasi backup
	config := BackupConfig{
		BaseDumpArgs:        s.Config.Backup.MysqlDumpArgs,
		OutputDir:           s.BackupOptions.OutputDirectory,
		CompressionType:     s.BackupOptions.Compression.Type,
		CompressionRequired: s.BackupOptions.Compression.Enabled,
		EncryptionEnabled:   s.BackupOptions.Encryption.Enabled,
	}

	// Log konfigurasi
	if config.EncryptionEnabled {
		s.Logger.Info("Enkripsi AES-256-GCM diaktifkan untuk backup (kompatibel dengan OpenSSL)")
	} else {
		s.Logger.Info("Enkripsi tidak diaktifkan, melewati langkah kunci enkripsi...")
	}

	if config.CompressionRequired {
		s.Logger.Infof("Kompresi %s diaktifkan (level: %s)", config.CompressionType, s.BackupOptions.Compression.Level)
	} else {
		s.Logger.Info("Kompresi tidak diaktifkan, melewati langkah kompresi...")
	}

	return config, nil
}

// ValidateOutput membuat direktori output jika belum ada
func (s *Service) ValidateOutput() error {

	// Membuat direktori output jika belum ada
	OutputDir, err := fs.CreateOutputDirs(s.BackupOptions.OutputDirectory, s.Config.Backup.Output.Structure.CreateSubdirs, s.Config.Backup.Output.Structure.Pattern, s.Config.General.ClientCode)
	if err != nil {
		return err
	}

	// Update struct dengan path yang sudah divalidasi
	s.BackupOptions.OutputDirectory = OutputDir
	return nil
}

// GenerateBackupFilename adalah helper untuk generate nama file backup
func (s *Service) GenerateBackupFilename(databaseName string) (string, error) {
	return fs.GenerateBackupFilename(
		s.Config.Backup.Output.Naming.Pattern,
		databaseName,
		"full",
		s.Config.Backup.Output.Naming.IncludeClientCode,
		s.Config.Backup.Output.Naming.IncludeHostname,
		s.Config.General.ClientCode,
		"",
	)
}
