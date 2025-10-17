package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime" // Diperlukan untuk konkurensi
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/ui"
	"strings"
	"sync" // Diperlukan untuk konkurensi
	"time"
)

// ExecuteBackup adalah satu-satunya entry point untuk menjalankan proses backup.
// Ia mengorkestrasi seluruh alur: setup, eksekusi, pengumpulan detail, dan pembuatan summary.
func (s *Service) ExecuteBackup(ctx context.Context, dbFiltered []string, backupMode string, collectDetails bool) error {
	// 1. Setup konfigurasi backup
	config, err := s.SetupBackupExecution()
	if err != nil {
		return fmt.Errorf("gagal setup backup execution: %w", err)
	}

	startTime := time.Now()
	var result backupResult

	// 2. (Opsional) Kumpulkan detail database jika diminta — lakukan sebelum backup
	// Detail database diperlukan untuk pengecekan disk space dan summary
	needDatabaseDetails := collectDetails || s.BackupOptions.DiskCheck

	if needDatabaseDetails {
		if len(dbFiltered) > 0 {
			// Pastikan nama database unik
			uniqueDBs := make(map[string]bool)
			for _, dbName := range dbFiltered {
				uniqueDBs[dbName] = true
			}
			dbNames := make([]string, 0, len(uniqueDBs))
			for dbName := range uniqueDBs {
				dbNames = append(dbNames, dbName)
			}

			s.Logger.Info("Mengumpulkan detail informasi database")
			var targetClient *database.Client
			targetClient, err = s.Client.ConnectToTargetDB(ctx)
			if err != nil {
				return fmt.Errorf("gagal koneksi ke target database: %w", err)
			}

			s.DatabaseDetail, err = targetClient.GetDatabaseDetails(ctx, dbNames, s.DBConfigInfo.ServerDBConnection.Host, s.DBConfigInfo.ServerDBConnection.Port)
			if err != nil {
				s.Logger.Warnf("Gagal mengumpulkan detail database: %v", err)
				return err
			} else {
				s.Logger.Infof("Berhasil mengumpulkan detail untuk %d database dari tabel database_details.", len(s.DatabaseDetail))
			}
		} else {
			s.Logger.Info("Tidak ada database untuk dikumpulkan detailnya sebelum backup.")
		}
	}

	// 3. Check Disk Space jika diaktifkan dan hitung estimasi
	var estimatesMap map[string]uint64 // Map database name ke estimasi ukuran

	// Periksa apakah pengecekan disk diaktifkan dan backup data tidak dikecualikan
	if s.BackupOptions.DiskCheck && !s.BackupOptions.Exclude.Data {
		var err error
		estimatesMap, err = s.checkDiskSpaceBeforeBackup(ctx, config, dbFiltered, backupMode)
		if err != nil {
			return fmt.Errorf("pengecekan ruang disk gagal: %w", err)
		}
	} else {
		s.Logger.Info("Pengecekan ruang disk tidak diaktifkan, melewati langkah ini...")
		estimatesMap = make(map[string]uint64)
	}

	ui.PrintSubHeader("Memulai Proses Backup")
	// 4. Lakukan backup berdasarkan mode
	if backupMode == "separate" {
		result = s.executeBackupSeparate(ctx, config, dbFiltered, estimatesMap)
	} else {
		result = s.executeBackupCombined(ctx, config, dbFiltered, estimatesMap)
	}

	// 5. Buat, simpan, dan tampilkan summary
	summary := s.CreateBackupSummary(backupMode, dbFiltered, result.successful, result.failed, startTime, result.errors)
	// Populate server version using the active client if available
	if s.Client != nil {
		if ver, err := s.Client.GetVersion(ctx); err == nil {
			summary.ServerInfo.Version = ver
		} else {
			s.Logger.Debugf("Gagal mengambil versi server dari client aktif: %v", err)
		}
	}

	if err := s.SaveSummaryToJSON(summary); err != nil {
		s.Logger.Errorf("Gagal menyimpan summary ke JSON: %v", err)
	}
	s.DisplaySummaryTable(summary)

	// 6. Kembalikan error jika ada kegagalan
	if len(result.failed) > 0 {
		var failedNames []string
		for _, failed := range result.failed {
			failedNames = append(failedNames, failed.DatabaseName)
		}
		return fmt.Errorf("beberapa database gagal di-backup: %v", failedNames)
	}

	return nil
}

// executeBackupSeparate melakukan backup dengan file terpisah per database secara paralel menggunakan worker pool.
func (s *Service) executeBackupSeparate(ctx context.Context, config BackupConfig, dbFiltered []string, estimatesMap map[string]uint64) backupResult {
	dbCount := len(dbFiltered)
	s.Logger.Infof("Total database yang akan di-backup: %d", dbCount)

	// Tentukan jumlah worker, idealnya sejumlah core CPU untuk efisiensi.
	numWorkers := runtime.NumCPU()
	if dbCount < numWorkers {
		numWorkers = dbCount // Tidak perlu lebih banyak worker dari jumlah pekerjaan
	}
	s.Logger.Infof("Menggunakan %d worker untuk backup paralel.", numWorkers)

	jobs := make(chan string, dbCount)
	results := make(chan jobResult, dbCount)
	var wg sync.WaitGroup

	// Hidupkan worker (goroutine)
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for dbName := range jobs {
				s.Logger.Infof("[Worker %d] Memproses database: %s", workerID, dbName)
				estimatedSize := estimatesMap[dbName]
				info, err := s.backupSingleDatabase(ctx, config, dbName, estimatedSize)
				results <- jobResult{info: info, err: err, dbName: dbName}
			}
		}(w)
	}

	// Kirim pekerjaan ke worker
	for _, dbName := range dbFiltered {
		jobs <- dbName
	}
	close(jobs)

	// Tunggu semua worker selesai
	wg.Wait()
	close(results)

	// Kumpulkan hasil dari semua worker
	var res backupResult
	for result := range results {
		if result.err != nil {
			errorMsg := fmt.Sprintf("Gagal backup database %s: %v", result.dbName, result.err)
			res.failed = append(res.failed, FailedDatabaseInfo{DatabaseName: result.dbName, Error: result.err.Error()})
			res.errors = append(res.errors, errorMsg)
			s.Logger.Error(errorMsg)
		} else {
			res.successful = append(res.successful, result.info)
			s.Logger.Infof("Database %s berhasil di-backup ke: %s", result.info.DatabaseName, result.info.OutputFile)
		}
	}

	s.Logger.Info("Proses backup database terpisah selesai.")
	s.Logger.Infof("Berhasil: %d database, Gagal: %d database", len(res.successful), len(res.failed))

	return res
}

// backupSingleDatabase menangani logika untuk membackup satu database.
func (s *Service) backupSingleDatabase(ctx context.Context, config BackupConfig, dbName string, estimatedSize uint64) (DatabaseBackupInfo, error) {
	startTime := time.Now()

	baseOutputFile, err := s.GenerateBackupFilename(dbName)
	if err != nil {
		return DatabaseBackupInfo{}, fmt.Errorf("gagal generate nama file: %w", err)
	}
	outputFile := s.addFileExtensions(baseOutputFile+".sql", config)
	fullOutputPath := filepath.Join(config.OutputDir, outputFile)

	mysqldumpArgs := s.buildMysqldumpArgs(config.BaseDumpArgs, nil, dbName)
	stderrOutput, err := s.executeMysqldumpWithPipe(ctx, mysqldumpArgs, fullOutputPath, config.CompressionRequired, config.CompressionType)

	// Tentukan status berdasarkan hasil eksekusi
	backupStatus := "success"
	var errorLogFile string

	if err != nil {
		// Fatal error - backup gagal total
		return DatabaseBackupInfo{}, err
	}

	// Jika ada stderr output (warnings/non-fatal errors), simpan ke file log
	if stderrOutput != "" {
		backupStatus = "success_with_warnings"
		errorLogFile = s.saveErrorLog(config.OutputDir, dbName, stderrOutput)
		s.Logger.Warnf("Database %s di-backup dengan warning (lihat: %s)", dbName, errorLogFile)
	}

	fileInfo, err := os.Stat(fullOutputPath)
	var fileSize int64
	if err == nil {
		fileSize = fileInfo.Size()
	}

	// Hitung compression ratio dan akurasi estimasi
	// CompressionRatio = BackupFileSize / OriginalDBSize
	// Misalnya: 0.15 = 15% dari ukuran asli (kompresi 85%)
	var compressionRatio float64
	var accuracyPct float64
	var estimatedSizeHuman string
	var originalDBSize int64
	var originalDBSizeHuman string

	// Dapatkan ukuran database asli dari detail
	if s.DatabaseDetail[dbName].SizeBytes > 0 {
		originalDBSize = s.DatabaseDetail[dbName].SizeBytes
		originalDBSizeHuman = s.DatabaseDetail[dbName].SizeHuman
		if fileSize > 0 {
			compressionRatio = float64(fileSize) / float64(originalDBSize)
		}
	}

	// Hitung akurasi estimasi disk space
	if estimatedSize > 0 && fileSize > 0 {
		accuracyPct = (float64(fileSize) / float64(estimatedSize)) * 100
		estimatedSizeHuman = s.formatFileSize(int64(estimatedSize))
	}

	return DatabaseBackupInfo{
		DatabaseName:        dbName,
		OutputFile:          fullOutputPath,
		FileSize:            fileSize,
		FileSizeHuman:       s.formatFileSize(fileSize),
		OriginalDBSize:      originalDBSize,
		OriginalDBSizeHuman: originalDBSizeHuman,
		CompressionRatio:    compressionRatio,
		EstimatedSize:       estimatedSize,
		EstimatedSizeHuman:  estimatedSizeHuman,
		AccuracyPercentage:  accuracyPct,
		Duration:            ui.FormatDuration(time.Since(startTime)),
		Status:              backupStatus,
		Warnings:            stderrOutput,
		ErrorLogFile:        errorLogFile,
	}, nil
}

// executeBackupCombined melakukan backup dengan satu file gabungan (sekuensial).
func (s *Service) executeBackupCombined(ctx context.Context, config BackupConfig, dbFiltered []string, estimatesMap map[string]uint64) backupResult {
	s.Logger.Info("Memulai proses backup...")

	var res backupResult
	backupStartTime := time.Now()

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

	mysqldumpArgs := s.buildMysqldumpArgs(config.BaseDumpArgs, dbFiltered, "")
	s.Logger.Debug("Direktori output: " + config.OutputDir)
	s.Logger.Debug("File output: " + fullOutputPath)

	stderrOutput, err := s.executeMysqldumpWithPipe(ctx, mysqldumpArgs, fullOutputPath, config.CompressionRequired, config.CompressionType)
	if err != nil {
		errorMsg := fmt.Errorf("gagal menjalankan mysqldump: %w", err)
		res.errors = append(res.errors, errorMsg.Error())
		for _, dbName := range dbFiltered {
			res.failed = append(res.failed, FailedDatabaseInfo{DatabaseName: dbName, Error: errorMsg.Error()})
		}
		return res
	}

	// Tentukan status berdasarkan stderr output
	backupStatus := "success"
	var errorLogFile string
	if stderrOutput != "" {
		backupStatus = "success_with_warnings"
		errorLogFile = s.saveErrorLog(config.OutputDir, "all_databases", stderrOutput)
		s.Logger.Warnf("Backup combined selesai dengan warning (lihat: %s)", errorLogFile)
	}

	backupDuration := time.Since(backupStartTime)
	fileInfo, err := os.Stat(fullOutputPath)
	var fileSize int64
	if err == nil {
		fileSize = fileInfo.Size()
	}

	// Hitung total estimasi dan total ukuran database asli untuk combined backup
	var totalEstimated uint64
	var totalOriginalSize int64
	for _, dbName := range dbFiltered {
		if est, ok := estimatesMap[dbName]; ok {
			totalEstimated += est
		}
		if detail, ok := s.DatabaseDetail[dbName]; ok {
			totalOriginalSize += detail.SizeBytes
		}
	}

	// Hitung compression ratio dan akurasi estimasi untuk combined backup
	// CompressionRatio = BackupFileSize / TotalOriginalDBSize
	// Accuracy = (BackupFileSize / EstimatedSize) * 100
	var compressionRatio float64
	var accuracyPct float64
	var estimatedSizeHuman string
	var originalDBSizeHuman string

	if totalOriginalSize > 0 && fileSize > 0 {
		compressionRatio = float64(fileSize) / float64(totalOriginalSize)
		originalDBSizeHuman = s.formatFileSize(totalOriginalSize)
	}

	if totalEstimated > 0 && fileSize > 0 {
		accuracyPct = (float64(fileSize) / float64(totalEstimated)) * 100
		estimatedSizeHuman = s.formatFileSize(int64(totalEstimated))
	}

	for _, dbName := range dbFiltered {
		res.successful = append(res.successful, DatabaseBackupInfo{
			DatabaseName:        dbName,
			OutputFile:          fullOutputPath,
			FileSize:            fileSize,
			FileSizeHuman:       s.formatFileSize(fileSize),
			OriginalDBSize:      totalOriginalSize,
			OriginalDBSizeHuman: originalDBSizeHuman,
			CompressionRatio:    compressionRatio,
			EstimatedSize:       totalEstimated,
			EstimatedSizeHuman:  estimatedSizeHuman,
			AccuracyPercentage:  accuracyPct,
			Duration:            ui.FormatDuration(backupDuration),
			Status:              backupStatus,
			Warnings:            stderrOutput,
			ErrorLogFile:        errorLogFile,
		})
	}

	s.Logger.Info("Proses backup semua database selesai.")
	s.Logger.Infof("File backup tersimpan di: %s", fullOutputPath)
	return res
}

// executeMysqldumpWithPipe menjalankan mysqldump dengan pipe untuk kompresi dan enkripsi.
// Mengembalikan error untuk fatal errors dan stderr output untuk warnings/non-fatal errors
func (s *Service) executeMysqldumpWithPipe(ctx context.Context, mysqldumpArgs []string, outputPath string, compressionRequired bool, compressionType string) (string, error) {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("gagal membuat file output: %w", err)
	}
	defer outputFile.Close()

	var writer io.Writer = outputFile
	var closers []io.Closer

	// Urutan layer: mysqldump -> Compression -> Encryption -> File
	if s.BackupOptions.Encryption.Enabled {
		encryptionKey := s.BackupOptions.Encryption.Key
		if encryptionKey == "" {
			resolvedKey, source, err := encrypt.ResolveEncryptionKey("")
			if err != nil {
				return "", fmt.Errorf("gagal mendapatkan kunci enkripsi: %w", err)
			}
			encryptionKey = resolvedKey
			s.Logger.Infof("Kunci enkripsi diperoleh dari: %s", source)
		}

		encryptingWriter, err := encrypt.NewEncryptingWriter(writer, []byte(encryptionKey))
		if err != nil {
			return "", fmt.Errorf("gagal membuat encrypting writer: %w", err)
		}
		closers = append(closers, encryptingWriter)
		writer = encryptingWriter
	}

	if compressionRequired {
		compressionConfig := compress.CompressionConfig{
			Type:  compress.CompressionType(compressionType),
			Level: compress.CompressionLevel(s.BackupOptions.Compression.Level),
		}
		compressingWriter, err := compress.NewCompressingWriter(writer, compressionConfig)
		if err != nil {
			return "", fmt.Errorf("gagal membuat compressing writer: %w", err)
		}
		closers = append(closers, compressingWriter)
		writer = compressingWriter
	}

	defer func() {
		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i].Close(); err != nil {
				s.Logger.Errorf("Error closing writer: %v", err)
			}
		}
	}()

	cmd := exec.CommandContext(ctx, "mysqldump", mysqldumpArgs...)
	cmd.Stdout = writer

	// Capture stderr untuk menangkap warnings dan errors
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	// logArgs := s.sanitizeArgsForLogging(mysqldumpArgs)
	// s.Logger.Infof("Command: mysqldump %s", strings.Join(logArgs, " "))

	if err := cmd.Run(); err != nil {
		stderrOutput := stderrBuf.String()
		// Cek apakah ini error fatal atau hanya warning
		if s.isFatalMysqldumpError(err, stderrOutput) {
			return stderrOutput, fmt.Errorf("mysqldump gagal: %w", err)
		}
		// Jika bukan fatal error, kembalikan stderr sebagai warning
		return stderrOutput, nil
	}

	return stderrBuf.String(), nil
}
