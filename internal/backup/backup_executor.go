package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime" // Diperlukan untuk konkurensi
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/ui"
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
	var databaseDetails map[string]structs.DatabaseDetail
	needDatabaseDetails := collectDetails || s.BackupOptions.DiskCheck

	if needDatabaseDetails {
		// Kumpulkan detail database sebelum backup
		if err := s.GetDBDetail(ctx, dbFiltered); err != nil {
			return fmt.Errorf("gagal mengumpulkan detail database: %w", err)
		}
	}

	// 3. Check Disk Space jika diaktifkan dan hitung estimasi
	var estimatesMap map[string]uint64 // Map database name ke estimasi ukuran
	if s.BackupOptions.DiskCheck {
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
		result = s.executeBackupSeparate(ctx, config, dbFiltered, estimatesMap, databaseDetails)
	} else {
		result = s.executeBackupCombined(ctx, config, dbFiltered, estimatesMap, databaseDetails)
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
func (s *Service) executeBackupSeparate(ctx context.Context, config BackupConfig, dbFiltered []string, estimatesMap map[string]uint64, databaseDetails map[string]structs.DatabaseDetail) backupResult {
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
				dbDetail := databaseDetails[dbName] // Ambil detail database
				info, err := s.backupSingleDatabase(ctx, config, dbName, estimatedSize, dbDetail)
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
func (s *Service) backupSingleDatabase(ctx context.Context, config BackupConfig, dbName string, estimatedSize uint64, dbDetail structs.DatabaseDetail) (DatabaseBackupInfo, error) {
	startTime := time.Now()

	baseOutputFile, err := s.GenerateBackupFilename(dbName)
	if err != nil {
		return DatabaseBackupInfo{}, fmt.Errorf("gagal generate nama file: %w", err)
	}
	outputFile := s.addFileExtensions(baseOutputFile+".sql", config)
	fullOutputPath := filepath.Join(config.OutputDir, outputFile)

	mysqldumpArgs := s.buildMysqldumpArgs(config.BaseDumpArgs, nil, dbName)
	if err := s.executeMysqldumpWithPipe(ctx, mysqldumpArgs, fullOutputPath, config.CompressionRequired, config.CompressionType); err != nil {
		return DatabaseBackupInfo{}, err
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
	if dbDetail.SizeBytes > 0 {
		originalDBSize = dbDetail.SizeBytes
		originalDBSizeHuman = dbDetail.SizeHuman
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
	}, nil
}

// executeBackupCombined melakukan backup dengan satu file gabungan (sekuensial).
func (s *Service) executeBackupCombined(ctx context.Context, config BackupConfig, dbFiltered []string, estimatesMap map[string]uint64, databaseDetails map[string]structs.DatabaseDetail) backupResult {
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

	if err := s.executeMysqldumpWithPipe(ctx, mysqldumpArgs, fullOutputPath, config.CompressionRequired, config.CompressionType); err != nil {
		errorMsg := fmt.Errorf("gagal menjalankan mysqldump: %w", err)
		res.errors = append(res.errors, errorMsg.Error())
		for _, dbName := range dbFiltered {
			res.failed = append(res.failed, FailedDatabaseInfo{DatabaseName: dbName, Error: errorMsg.Error()})
		}
		return res
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
		if detail, ok := databaseDetails[dbName]; ok {
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
		})
	}

	s.Logger.Info("Proses backup semua database selesai.")
	s.Logger.Infof("File backup tersimpan di: %s", fullOutputPath)
	return res
}

// executeMysqldumpWithPipe menjalankan mysqldump dengan pipe untuk kompresi dan enkripsi.
func (s *Service) executeMysqldumpWithPipe(ctx context.Context, mysqldumpArgs []string, outputPath string, compressionRequired bool, compressionType string) error {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("gagal membuat file output: %w", err)
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

	defer func() {
		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i].Close(); err != nil {
				s.Logger.Errorf("Error closing writer: %v", err)
			}
		}
	}()

	cmd := exec.CommandContext(ctx, "mysqldump", mysqldumpArgs...)
	cmd.Stdout = writer
	cmd.Stderr = os.Stderr

	// logArgs := s.sanitizeArgsForLogging(mysqldumpArgs)
	// s.Logger.Infof("Command: mysqldump %s", strings.Join(logArgs, " "))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysqldump gagal: %w", err)
	}

	return nil
}
