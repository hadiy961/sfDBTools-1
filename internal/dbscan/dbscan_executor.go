// File : internal/dbscan/dbscan_executor.go
// Deskripsi : Eksekutor utama untuk database scanning dan menyimpan hasil ke database
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 16 Oktober 2025

package dbscan

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
)

// ExecuteScanInBackground menjalankan scanning tanpa UI output (pure logging)
// Ini adalah "background" mode dalam artian tidak ada interaksi UI, bukan goroutine
// Process tetap berjalan sampai selesai, cocok untuk cron job atau automation
func (s *Service) ExecuteScanInBackground(ctx context.Context, config ScanEntryConfig) error {
	// Generate unique scan ID
	scanID := fmt.Sprintf("scan_%s", time.Now().Format("20060102_150405"))

	s.Logger.Infof("[%s] ========================================", scanID)
	s.Logger.Infof("[%s] DATABASE SCANNING - BACKGROUND MODE", scanID)
	s.Logger.Infof("[%s] ========================================", scanID)
	s.Logger.Infof("[%s] Memulai background scanning...", scanID)

	// Setup connections
	sourceClient, targetClient, dbFiltered, cleanup, err := s.setupScanConnections(ctx, "", false)
	if err != nil {
		s.Logger.Errorf("[%s] Gagal setup session: %v", scanID, err)
		return err
	}
	defer cleanup()

	// Ensure logs dir, create lockfile and pid file, and acquire exclusive lock (flock)
	logDir := filepath.Join("logs", "dbscan")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		s.Logger.Warnf("[%s] Gagal membuat log dir: %v", scanID, err)
	}

	lockFile := filepath.Join(logDir, "dbscan_background.lock")
	pidFile := filepath.Join(logDir, "dbscan_background.pid")

	// Open (or create) lock file
	lf, err := os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		s.Logger.Warnf("[%s] Gagal membuka lock file %s: %v", scanID, lockFile, err)
	} else {
		// Try to acquire exclusive non-blocking lock
		if err := syscall.Flock(int(lf.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
			// Couldn't acquire lock -> another process likely running
			lf.Close()
			s.Logger.Warnf("[%s] Tidak dapat memperoleh lock, proses background lain mungkin sedang berjalan (lockfile=%s)", scanID, lockFile)
			return fmt.Errorf("background process sudah berjalan (lockfile=%s)", lockFile)
		}
		s.Logger.Infof("[%s] Berhasil memperoleh lock: %s", scanID, lockFile)
	}

	// Write own PID to pid file
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
		s.Logger.Warnf("[%s] Gagal menulis pid file: %v", scanID, err)
	} else {
		s.Logger.Infof("[%s] Menulis PID file: %s", scanID, pidFile)
	}

	// Create cancellable context to support graceful shutdown
	runCtx, cancel := context.WithCancel(ctx)

	// Setup signal handler to cancel context (graceful shutdown)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		s.Logger.Warnf("[%s] Menerima sinyal %s - memulai graceful shutdown", scanID, sig.String())
		cancel()
	}()

	// Cleanup: remove pidfile, release flock, close lock file, stop signal notifications
	defer func() {
		// Remove pid file
		if err := os.Remove(pidFile); err != nil {
			s.Logger.Warnf("[%s] Gagal menghapus pid file %s: %v", scanID, pidFile, err)
		} else {
			s.Logger.Infof("[%s] PID file %s dihapus", scanID, pidFile)
		}

		// Release flock and close file
		if lf != nil {
			if err := syscall.Flock(int(lf.Fd()), syscall.LOCK_UN); err != nil {
				s.Logger.Warnf("[%s] Gagal melepaskan lock %s: %v", scanID, lockFile, err)
			} else {
				s.Logger.Infof("[%s] Lock dilepas: %s", scanID, lockFile)
			}
			lf.Close()
		}

		signal.Stop(sigs)
		cancel()
	}()

	// Lakukan scanning dengan background mode (pure logging)
	s.Logger.Infof("[%s] Scanning %d database...", scanID, len(dbFiltered))
	result, err := s.ExecuteScan(runCtx, sourceClient, targetClient, dbFiltered, true)
	if err != nil {
		s.Logger.Errorf("[%s] Scanning gagal: %v", scanID, err)
		return err
	}

	// Log hasil
	s.Logger.Infof("[%s] ========================================", scanID)
	s.Logger.Infof("[%s] HASIL SCANNING", scanID)
	s.Logger.Infof("[%s] ========================================", scanID)
	s.Logger.Infof("[%s] Total Database  : %d", scanID, result.TotalDatabases)
	s.Logger.Infof("[%s] Berhasil        : %d", scanID, result.SuccessCount)
	s.Logger.Infof("[%s] Gagal           : %d", scanID, result.FailedCount)
	s.Logger.Infof("[%s] Durasi          : %s", scanID, result.Duration)

	if len(result.Errors) > 0 {
		s.Logger.Warnf("[%s] Terdapat %d error saat scanning:", scanID, len(result.Errors))
		for i, errMsg := range result.Errors {
			s.Logger.Warnf("[%s]   %d. %s", scanID, i+1, errMsg)
		}
	}

	s.Logger.Infof("[%s] Background scanning selesai dengan sukses.", scanID)
	s.Logger.Infof("[%s] ========================================", scanID)

	return nil
}

// ExecuteScan melakukan scanning database dan menyimpan hasilnya
// Parameter isBackground menentukan apakah output menggunakan logger (true) atau UI (false)
func (s *Service) ExecuteScan(ctx context.Context, sourceClient *database.Client, targetClient *database.Client, dbNames []string, isBackground bool) (*ScanResult, error) {
	startTime := time.Now()

	if isBackground {
		s.Logger.Info("Memulai proses scanning database...")
		s.Logger.Infof("Total database yang akan di-scan: %d", len(dbNames))
	}
	ui.PrintSubHeader("Memulai Proses Scanning Database")

	if len(dbNames) == 0 {
		return nil, fmt.Errorf("tidak ada database untuk di-scan")
	}

	// Get Original max_statement_time
	s.Logger.Info("Mendapatkan nilai max_statement_time awal...")
	originalMaxStatementTime, err := sourceClient.GetMaxStatementsTime(ctx)
	if err != nil {
		s.Logger.Warn("Gagal mendapatkan max_statement_time awal: " + err.Error())
	} else {
		s.Logger.Infof("Original max_statement_time: %f detik", originalMaxStatementTime)
	}

	// Set max_statement_time ke 0 (no limit) untuk sesi ini
	s.Logger.Info("Mengatur max_statement_time ke 0 detik untuk sesi ini...")
	if err := sourceClient.SetMaxStatementsTime(ctx, 0); err != nil {
		s.Logger.Warn("Gagal mengatur max_statement_time: " + err.Error())
	}

	// Verifikasi nilai baru
	currentMaxStatementTime, err := sourceClient.GetMaxStatementsTime(ctx)
	if err != nil {
		s.Logger.Warn("Gagal mendapatkan max_statement_time saat ini: " + err.Error())
	} else {
		s.Logger.Debugf("Current max_statement_time: %f detik", currentMaxStatementTime)
	}

	// Mulai scanning
	s.Logger.Info("Memulai pengumpulan detail database...")

	// Collect database details
	detailsMap := database.CollectDatabaseDetails(ctx, sourceClient, dbNames, s.Logger)

	// Kembalikan max_statement_time ke nilai awal
	if originalMaxStatementTime > 0 {
		s.Logger.Info("Mengembalikan max_statement_time ke nilai awal...")
		if err := sourceClient.SetMaxStatementsTime(ctx, originalMaxStatementTime); err != nil {
			s.Logger.Warn("Gagal mengembalikan max_statement_time: " + err.Error())
		} else {
			s.Logger.Info("max_statement_time berhasil dikembalikan.")
		}
	}

	// Simpan ke database jika diminta
	successCount := 0
	failedCount := 0
	var errors []string
	totalToSave := len(detailsMap)

	if s.ScanOptions.SaveToDB && targetClient != nil {
		if isBackground {
			s.Logger.Infof("Menyimpan hasil scan ke database target (%d database)...", totalToSave)
		} else {
			ui.PrintInfo(fmt.Sprintf("Menyimpan hasil scan ke database target (%d database)...", totalToSave))
		}

		processedCount := 0
		lastLoggedPercent := 0

		for dbName, detail := range detailsMap {
			processedCount++
			err := s.SaveDatabaseDetail(ctx, targetClient, detail)

			if err != nil {
				if isBackground {
					s.Logger.Errorf("[%d/%d] ✗ Gagal menyimpan database %s: %v", processedCount, totalToSave, dbName, err)
				}
				errors = append(errors, fmt.Sprintf("%s: %v", dbName, err))
				failedCount++
			} else {
				successCount++
			}

			// Log milestone percentages (every 25%)
			if isBackground {
				currentPercent := (processedCount * 100) / totalToSave
				if currentPercent >= lastLoggedPercent+25 && currentPercent != 100 {
					s.Logger.Infof("Progress penyimpanan: %d%% selesai (%d/%d database)", currentPercent, processedCount, totalToSave)
					lastLoggedPercent = currentPercent
				}
			}
		}

		if isBackground {
			s.Logger.Infof("Penyimpanan selesai 100%% - Berhasil: %d, Gagal: %d", successCount, failedCount)
		}
	} else {
		successCount = len(detailsMap)
	}

	// Tampilkan hasil jika diminta (hanya untuk foreground)
	if s.ScanOptions.DisplayResults && !isBackground {
		s.DisplayDetailResults(detailsMap)
	}

	// Log detail untuk background mode
	if isBackground && s.ScanOptions.DisplayResults {
		s.LogDetailResults(detailsMap)
	}

	duration := time.Since(startTime)

	return &ScanResult{
		TotalDatabases: len(dbNames),
		SuccessCount:   successCount,
		FailedCount:    failedCount,
		Duration:       duration.String(),
		Errors:         errors,
	}, nil
}

// SaveDatabaseDetail menyimpan detail database ke tabel database_details menggunakan stored procedure
func (s *Service) SaveDatabaseDetail(ctx context.Context, client *database.Client, detail database.DatabaseDetailInfo) error {
	// Parse collection time
	collectionTime, err := time.Parse("2006-01-02 15:04:05", detail.CollectionTime)
	if err != nil {
		collectionTime = time.Now()
	}

	// Ambil server info dari ScanOptions
	serverHost := s.ScanOptions.DBConfig.ServerDBConnection.Host
	serverPort := s.ScanOptions.DBConfig.ServerDBConnection.Port

	// Prepare error message (NULL if no error)
	var errorMsg *string
	if detail.Error != "" {
		errorMsg = &detail.Error
	}

	// Call stored procedure
	query := `CALL sp_insert_database_detail(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = client.DB().ExecContext(ctx, query,
		detail.DatabaseName,
		serverHost,
		serverPort,
		detail.SizeBytes,
		detail.SizeHuman,
		detail.TableCount,
		detail.ProcedureCount,
		detail.FunctionCount,
		detail.ViewCount,
		detail.UserGrantCount,
		collectionTime,
		errorMsg,
		0,
	)

	return err
}
