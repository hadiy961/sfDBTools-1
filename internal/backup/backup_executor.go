package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/ui"
	"strings"
	"time"
)

// ExecuteBackup adalah satu-satunya entry point untuk menjalankan proses backup.
// Ia mengorkestrasi seluruh alur: setup, eksekusi, pengumpulan detail (opsional), dan pembuatan summary.
// REFACTOR: Mengonsolidasikan semua fungsi Execute... menjadi satu entry point yang jelas.
func (s *Service) ExecuteBackup(ctx context.Context, client *database.Client, dbFiltered []string, backupMode string, collectDetails bool) error {
	// 1. Setup konfigurasi backup
	config, err := s.SetupBackupExecution()
	if err != nil {
		return fmt.Errorf("gagal setup backup execution: %w", err)
	}

	startTime := time.Now()
	var result backupResult

	// 2. Lakukan backup berdasarkan mode
	if backupMode == "separate" {
		result = s.executeBackupSeparate(ctx, config, dbFiltered)
	} else {
		result = s.executeBackupCombined(ctx, config, dbFiltered)
	}

	// 3. (Opsional) Kumpulkan detail database jika diminta dan ada backup yang berhasil
	var databaseDetails map[string]DatabaseDetailInfo
	if collectDetails && len(result.successful) > 0 {
		uniqueDBs := make(map[string]bool)
		for _, db := range result.successful {
			uniqueDBs[db.DatabaseName] = true
		}
		dbNames := make([]string, 0, len(uniqueDBs))
		for dbName := range uniqueDBs {
			dbNames = append(dbNames, dbName)
		}

		s.Logger.Info("Mengumpulkan detail informasi database...")
		databaseDetails = s.CollectDatabaseDetails(ctx, client, dbNames)
	}

	// 4. Buat, simpan, dan tampilkan summary
	summary := s.CreateBackupSummaryWithDetails(backupMode, dbFiltered, result.successful, result.failed, startTime, result.errors, databaseDetails)
	if err := s.SaveSummaryToJSON(summary); err != nil {
		s.Logger.Errorf("Gagal menyimpan summary ke JSON: %v", err)
	}
	s.DisplaySummaryTable(summary)

	// 5. Kembalikan error jika ada database yang gagal di-backup
	if len(result.failed) > 0 {
		var failedNames []string
		for _, failed := range result.failed {
			failedNames = append(failedNames, failed.DatabaseName)
		}
		return fmt.Errorf("beberapa database gagal di-backup: %v", failedNames)
	}

	return nil
}

// executeBackupSeparate melakukan backup dengan file terpisah per database.
// REFACTOR: Disederhanakan, tidak lagi menerima pointer slice, tapi mengembalikan struct backupResult.
func (s *Service) executeBackupSeparate(ctx context.Context, config BackupConfig, dbFiltered []string) backupResult {
	ui.PrintSubHeader("Proses Backup Database Terpisah")
	s.Logger.Infof("Total database yang akan di-backup: %d", len(dbFiltered))

	var res backupResult

	for i, dbName := range dbFiltered {
		s.Logger.Infof("Memproses database [%d/%d]: %s", i+1, len(dbFiltered), dbName)

		// REFACTOR: Logika untuk mem-backup satu DB diekstrak ke fungsi sendiri.
		info, err := s.backupSingleDatabase(ctx, config, dbName)
		if err != nil {
			errorMsg := fmt.Sprintf("Gagal backup database %s: %v", dbName, err)
			res.failed = append(res.failed, FailedDatabaseInfo{DatabaseName: dbName, Error: err.Error()})
			res.errors = append(res.errors, errorMsg)
			s.Logger.Error(errorMsg)
			continue
		}

		res.successful = append(res.successful, info)
		s.Logger.Infof("Database %s berhasil di-backup ke: %s", dbName, info.OutputFile)
	}

	s.Logger.Info("Proses backup database terpisah selesai.")
	s.Logger.Infof("Berhasil: %d database, Gagal: %d database", len(res.successful), len(res.failed))

	return res
}

// backupSingleDatabase menangani logika untuk membackup satu database.
// REFACTOR: Fungsi baru hasil ekstraksi untuk meningkatkan keterbacaan dan reusability.
func (s *Service) backupSingleDatabase(ctx context.Context, config BackupConfig, dbName string) (DatabaseBackupInfo, error) {
	startTime := time.Now()

	// Generate nama file
	baseOutputFile, err := s.GenerateBackupFilename(dbName)
	if err != nil {
		return DatabaseBackupInfo{}, fmt.Errorf("gagal generate nama file: %w", err)
	}
	outputFile := s.addFileExtensions(baseOutputFile+".sql", config)
	fullOutputPath := filepath.Join(config.OutputDir, outputFile)

	// Siapkan dan jalankan mysqldump
	mysqldumpArgs := s.buildMysqldumpArgs(config.BaseDumpArgs, nil, dbName)

	if err := s.executeMysqldumpWithPipe(ctx, mysqldumpArgs, fullOutputPath, config.CompressionRequired, config.CompressionType); err != nil {
		return DatabaseBackupInfo{}, err
	}

	// Kumpulkan informasi hasil backup
	fileInfo, err := os.Stat(fullOutputPath)
	var fileSize int64
	if err == nil {
		fileSize = fileInfo.Size()
	}

	return DatabaseBackupInfo{
		DatabaseName:  dbName,
		OutputFile:    fullOutputPath,
		FileSize:      fileSize,
		FileSizeHuman: s.FormatFileSize(fileSize),
		Duration:      s.FormatDuration(time.Since(startTime)),
	}, nil
}

// executeBackupCombined melakukan backup dengan satu file gabungan.
// REFACTOR: Disederhanakan, mengembalikan struct backupResult.
func (s *Service) executeBackupCombined(ctx context.Context, config BackupConfig, dbFiltered []string) backupResult {
	ui.PrintSubHeader("Proses Backup Database Gabungan")
	s.Logger.Info("Memulai proses backup...")

	var res backupResult
	backupStartTime := time.Now()

	// Generate nama file
	baseOutputFile, err := s.GenerateBackupFilename("all_databases")
	if err != nil {
		errorMsg := fmt.Errorf("gagal generate nama file backup: %w", err)
		res.errors = append(res.errors, errorMsg.Error())
		for _, dbName := range dbFiltered {
			res.failed = append(res.failed, FailedDatabaseInfo{DatabaseName: dbName, Error: errorMsg.Error()})
		}
		return res
	}

	outputFile := s.addFileExtensions(baseOutputFile+".sql", config)
	fullOutputPath := filepath.Join(config.OutputDir, outputFile)

	// Siapkan dan jalankan mysqldump
	mysqldumpArgs := s.buildMysqldumpArgs(config.BaseDumpArgs, dbFiltered, "")
	s.Logger.Debug("Direktori output: " + config.OutputDir)
	s.Logger.Debug("File output: " + fullOutputPath)

	if err := s.executeMysqldumpWithPipe(ctx, mysqldumpArgs, fullOutputPath, config.CompressionRequired, config.CompressionType); err != nil {
		errorMsg := fmt.Errorf("gagal menjalankan mysqldump: %w", err)
		res.errors = append(res.errors, errorMsg.Error())
		for _, dbName := range dbFiltered {
			res.failed = append(res.failed, FailedDatabaseInfo{DatabaseName: dbName, Error: errorMsg.Error()})
		}
		return res
	}

	// Kumpulkan informasi hasil backup
	backupDuration := time.Since(backupStartTime)
	fileInfo, err := os.Stat(fullOutputPath)
	var fileSize int64
	if err == nil {
		fileSize = fileInfo.Size()
	}

	// Buat entri untuk setiap database yang berhasil di-backup dalam satu file
	for _, dbName := range dbFiltered {
		res.successful = append(res.successful, DatabaseBackupInfo{
			DatabaseName:  dbName,
			OutputFile:    fullOutputPath,
			FileSize:      fileSize, // Ukuran file sama untuk semua DB dalam mode ini
			FileSizeHuman: s.FormatFileSize(fileSize),
			Duration:      s.FormatDuration(backupDuration), // Durasi juga sama
		})
	}

	s.Logger.Info("Proses backup semua database selesai.")
	s.Logger.Infof("File backup tersimpan di: %s", fullOutputPath)
	return res
}

// executeMysqldumpWithPipe menjalankan mysqldump dengan pipe ke file output.
// (Fungsi ini sudah cukup baik dan tidak memerlukan perubahan signifikan).
func (s *Service) executeMysqldumpWithPipe(ctx context.Context, mysqldumpArgs []string, outputPath string, compressionRequired bool, compressionType string) error {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("gagal membuat file output: %w", err)
	}
	defer outputFile.Close()

	var writer io.Writer = outputFile
	var closers []io.Closer

	// Setup enkripsi jika diperlukan (layer pertama: paling dalam)
	// Urutan layer: File -> Encryption -> Compression -> mysqldump
	if s.BackupOptions.Encryption.Enabled {
		encryptionKey := s.BackupOptions.Encryption.Key

		if encryptionKey == "" {
			resolvedKey, source, err := encrypt.ResolveEncryptionKey("")
			if err != nil {
				return fmt.Errorf("gagal mendapatkan kunci enkripsi: %w", err)
			}
			encryptionKey = resolvedKey
			s.Logger.Infof("Kunci enkripsi diperoleh dari: %s", source)
		}

		encryptingWriter, err := encrypt.NewEncryptingWriter(writer, []byte(encryptionKey))
		if err != nil {
			return fmt.Errorf("gagal membuat encrypting writer: %w", err)
		}
		closers = append(closers, encryptingWriter)
		writer = encryptingWriter

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

	}

	// Defer semua closers dalam urutan terbalik
	defer func() {
		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i].Close(); err != nil {
				s.Logger.Errorf("Error closing writer: %v", err)
			}
		}
	}()

	// Buat dan jalankan command mysqldump
	cmd := exec.CommandContext(ctx, "mysqldump", mysqldumpArgs...)
	cmd.Stdout = writer
	cmd.Stderr = os.Stderr

	logArgs := s.sanitizeArgsForLogging(mysqldumpArgs)
	s.Logger.Infof("Command: mysqldump %s", strings.Join(logArgs, " "))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysqldump gagal: %w", err)
	}

	return nil
}
