package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/ui"
	"strings"
)

// ExecuteBackup adalah fungsi general untuk melakukan backup database
// backupMode: "separate" untuk file terpisah per database, "combined" untuk satu file gabungan
func (s *Service) ExecuteBackup(ctx context.Context, dbFiltered []string, backupMode string) error {
	// Setup konfigurasi backup (cleanup, validate, konfigurasi)
	config, err := s.SetupBackupExecution()
	if err != nil {
		return fmt.Errorf("gagal setup backup execution: %w", err)
	}

	if backupMode == "separate" {
		return s.executeBackupSeparate(ctx, config, dbFiltered)
	} else {
		return s.executeBackupCombined(ctx, config, dbFiltered)
	}
}

// executeBackupSeparate melakukan backup dengan file terpisah per database
func (s *Service) executeBackupSeparate(ctx context.Context, config BackupConfig, dbFiltered []string) error {
	ui.PrintSubHeader("Proses Backup Database Terpisah")

	s.Logger.Info("Memulai proses backup database secara terpisah...")
	s.Logger.Infof("Total database yang akan di-backup: %d", len(dbFiltered))

	// Counter untuk tracking backup yang berhasil dan gagal
	successCount := 0
	failedDatabases := []string{}

	// Loop melalui setiap database untuk backup terpisah
	for i, dbName := range dbFiltered {
		s.Logger.Infof("Memproses database [%d/%d]: %s", i+1, len(dbFiltered), dbName)

		// Generate nama file untuk database ini
		outputFile, err := s.GenerateBackupFilename(dbName)
		if err != nil {
			s.Logger.Errorf("Gagal generate nama file untuk database %s: %v", dbName, err)
			failedDatabases = append(failedDatabases, dbName)
			continue
		}
		// Buat nama file output untuk database ini
		outputFile = outputFile + ".sql"

		// Tambahkan ekstensi kompresi dan enkripsi
		outputFile = s.addFileExtensions(outputFile, config)
		fullOutputPath := filepath.Join(config.OutputDir, outputFile)

		// Siapkan argumen mysqldump untuk database tunggal ini
		mysqldumpArgs := s.buildMysqldumpArgs(config.BaseDumpArgs, dbFiltered, dbName)

		s.Logger.Debug("Direktori output: " + config.OutputDir)
		s.Logger.Debug("File output: " + fullOutputPath)

		// Jalankan mysqldump dengan pipe ke output file
		if err := s.executeMysqldumpWithPipe(ctx, mysqldumpArgs, fullOutputPath, config.CompressionRequired, config.CompressionType); err != nil {
			s.Logger.Errorf("Gagal backup database %s: %v", dbName, err)
			failedDatabases = append(failedDatabases, dbName)
			continue
		}

		s.Logger.Infof("Database %s berhasil di-backup ke: %s", dbName, fullOutputPath)
		successCount++
	}

	// Laporan hasil backup
	s.Logger.Info("Proses backup database terpisah selesai.")
	s.Logger.Infof("Berhasil: %d database, Gagal: %d database", successCount, len(failedDatabases))

	if len(failedDatabases) > 0 {
		s.Logger.Warnf("Database yang gagal di-backup: %v", failedDatabases)
		return fmt.Errorf("beberapa database gagal di-backup: %v", failedDatabases)
	}

	return nil
}

// executeBackupCombined melakukan backup dengan satu file gabungan
func (s *Service) executeBackupCombined(ctx context.Context, config BackupConfig, dbFiltered []string) error {
	ui.PrintSubHeader("Proses Backup Database")

	s.Logger.Info("Memulai proses backup database...")

	// Generate nama file untuk backup all databases
	baseOutputFile, err := s.GenerateBackupFilename("all_databases")
	if err != nil {
		return fmt.Errorf("gagal generate nama file backup: %w", err)
	}

	// Buat nama file output
	outputFile := baseOutputFile + ".sql"

	// Tambahkan ekstensi kompresi dan enkripsi
	outputFile = s.addFileExtensions(outputFile, config)
	fullOutputPath := filepath.Join(config.OutputDir, outputFile)

	// Siapkan argumen mysqldump dengan kredensial database
	mysqldumpArgs := s.buildMysqldumpArgs(config.BaseDumpArgs, dbFiltered, "")

	s.Logger.Debug("Direktori output: " + config.OutputDir)
	s.Logger.Debug("File output: " + fullOutputPath)

	// Jalankan mysqldump dengan pipe ke output file
	if err := s.executeMysqldumpWithPipe(ctx, mysqldumpArgs, fullOutputPath, config.CompressionRequired, config.CompressionType); err != nil {
		return fmt.Errorf("gagal menjalankan mysqldump: %w", err)
	}

	s.Logger.Info("Proses backup semua database selesai.")
	s.Logger.Infof("File backup tersimpan di: %s", fullOutputPath)

	return nil
}

// executeMysqldumpWithPipe menjalankan mysqldump dengan pipe ke file output
func (s *Service) executeMysqldumpWithPipe(ctx context.Context, mysqldumpArgs []string, outputPath string, compressionRequired bool, compressionType string) error {
	// Buat file output
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("gagal membuat file output: %w", err)
	}
	defer outputFile.Close()

	var writer io.Writer = outputFile
	var closers []io.Closer

	// Setup enkripsi jika diperlukan (layer pertama: paling dalam)
	// Urutan layer: File -> Encryption -> Compression -> mysqldump
	encryptionEnabled := s.BackupOptions.Encryption.Enabled
	if encryptionEnabled {
		encryptionKey := s.BackupOptions.Encryption.Key

		// Resolve encryption key jika belum ada
		if encryptionKey == "" {
			resolvedKey, source, err := encrypt.ResolveEncryptionKey("")
			if err != nil {
				return fmt.Errorf("gagal mendapatkan kunci enkripsi: %w", err)
			}
			encryptionKey = resolvedKey
			s.Logger.Infof("Kunci enkripsi diperoleh dari: %s", source)
		}

		// Buat encrypting writer dengan format kompatibel OpenSSL
		encryptingWriter, err := encrypt.NewEncryptingWriter(writer, []byte(encryptionKey))
		if err != nil {
			return fmt.Errorf("gagal membuat encrypting writer: %w", err)
		}
		closers = append(closers, encryptingWriter)
		writer = encryptingWriter

		s.Logger.Info("Enkripsi AES-256-GCM diaktifkan untuk backup (kompatibel dengan OpenSSL)")
	}

	// Setup kompresi jika diperlukan (layer kedua: di atas enkripsi)
	if compressionRequired {
		compressionConfig := compress.CompressionConfig{
			Type:  compress.CompressionType(compressionType),
			Level: compress.CompressionLevel(s.BackupOptions.Compression.Level),
		}

		compressingWriter, err := compress.NewCompressingWriter(writer, compressionConfig)
		if err != nil {
			return fmt.Errorf("gagal membuat compressing writer: %w", err)
		}
		closers = append(closers, compressingWriter)
		writer = compressingWriter

		s.Logger.Infof("Kompresi %s diaktifkan (level: %s)", compressionType, s.BackupOptions.Compression.Level)
	}

	// Defer semua closers dalam urutan terbalik
	defer func() {
		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i].Close(); err != nil {
				s.Logger.Errorf("Error closing writer: %v", err)
			}
		}
	}()

	// Buat command mysqldump
	cmd := exec.CommandContext(ctx, "mysqldump", mysqldumpArgs...)

	// Set stdout ke writer (file atau compressed writer)
	cmd.Stdout = writer
	cmd.Stderr = os.Stderr

	s.Logger.Info("Menjalankan mysqldump...")
	// Log command dengan password disembunyikan untuk keamanan
	logArgs := s.sanitizeArgsForLogging(mysqldumpArgs)
	s.Logger.Debugf("Command: mysqldump %s", strings.Join(logArgs, " "))

	// Jalankan command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysqldump gagal: %w", err)
	}

	s.Logger.Info("Mysqldump berhasil dijalankan")
	return nil
}
