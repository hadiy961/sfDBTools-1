// File : internal/dbscan/dbscan_executor.go
// Deskripsi : Eksekutor utama untuk database scanning dan menyimpan hasil ke database
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 16 Oktober 2025

package dbscan

import (
	"context"
	"fmt"
	"time"

	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
)

// ExecuteScanBackground menjalankan scanning di background (asynchronous)
func (s *Service) ExecuteScanBackground(ctx context.Context, config ScanEntryConfig) error {
	ui.PrintHeader("DATABASE SCANNING - BACKGROUND MODE")
	ui.PrintInfo("Scanning akan dijalankan di background...")

	// Generate unique scan ID
	scanID := fmt.Sprintf("scan_%s", time.Now().Format("20060102_150405"))

	ui.PrintInfo(fmt.Sprintf("Scan ID: %s", ui.ColorText(scanID, ui.ColorCyan)))
	ui.PrintInfo("Logs akan ditulis ke logger. Gunakan 'tail -f' untuk monitoring progress.")

	// Jalankan scanning di goroutine
	go func() {
		s.Logger.Infof("[%s] Memulai background scanning...", scanID)

		// Setup connections
		sourceClient, targetClient, dbFiltered, cleanup, err := s.setupScanConnections(ctx, "", false)
		if err != nil {
			s.Logger.Errorf("[%s] Gagal setup session: %v", scanID, err)
			return
		}
		defer cleanup()

		// Lakukan scanning
		s.Logger.Infof("[%s] Scanning %d database...", scanID, len(dbFiltered))
		result, err := s.ExecuteScan(ctx, sourceClient, targetClient, dbFiltered, true)
		if err != nil {
			s.Logger.Errorf("[%s] Scanning gagal: %v", scanID, err)
			return
		}

		// Log hasil
		s.Logger.Infof("[%s] Scanning selesai - Total: %d, Success: %d, Failed: %d, Duration: %s",
			scanID, result.TotalDatabases, result.SuccessCount, result.FailedCount, result.Duration)

		if len(result.Errors) > 0 {
			s.Logger.Warnf("[%s] Terdapat %d error saat scanning", scanID, len(result.Errors))
			for _, errMsg := range result.Errors {
				s.Logger.Warnf("[%s] Error: %s", scanID, errMsg)
			}
		}
	}()

	ui.PrintSuccess(fmt.Sprintf("Background scanning dimulai dengan ID: %s", scanID))
	ui.PrintInfo("Process akan berjalan di background. Check logs untuk monitoring.")

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

	if len(dbNames) == 0 {
		return nil, fmt.Errorf("tidak ada database untuk di-scan")
	}

	// Collect database details
	detailsMap := database.CollectDatabaseDetails(ctx, sourceClient, dbNames, s.Logger)

	// Simpan ke database jika diminta
	successCount := 0
	failedCount := 0
	var errors []string

	if s.ScanOptions.SaveToDB && targetClient != nil {
		if isBackground {
			s.Logger.Info("Menyimpan hasil scan ke database...")
		} else {
			ui.PrintInfo("Menyimpan hasil scan ke database...")
		}

		for dbName, detail := range detailsMap {
			err := s.SaveDatabaseDetail(ctx, targetClient, detail)
			if err != nil {
				s.Logger.Errorf("Gagal menyimpan detail database %s: %v", dbName, err)
				errors = append(errors, fmt.Sprintf("%s: %v", dbName, err))
				failedCount++
			} else {
				successCount++
			}
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
