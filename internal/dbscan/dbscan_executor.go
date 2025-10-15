// File : internal/dbscan/dbscan_executor.go
// Deskripsi : Eksekutor utama untuk database scanning dan menyimpan hasil ke database
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025

package dbscan

import (
	"context"
	"fmt"
	"time"

	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
)

// ExecuteScanCommand adalah entry point untuk database scan
func (s *Service) ExecuteScanCommand(config ScanEntryConfig) error {
	ctx := context.Background()

	// Jika background mode, jalankan secara asynchronous
	if s.ScanOptions.Background {
		return s.ExecuteScanBackground(ctx, config)
	}

	// Setup session (koneksi database source)
	sourceClient, dbFiltered, err := s.PrepareScanSession(ctx, config.HeaderTitle, config.ShowOptions)
	if err != nil {
		return err
	}
	defer sourceClient.Close()

	// Koneksi ke target database untuk menyimpan hasil scan
	var targetClient *database.Client
	if s.ScanOptions.SaveToDB {
		targetClient, err = s.ConnectToTargetDB(ctx)
		if err != nil {
			s.Logger.Warn("Gagal koneksi ke target database, hasil scan tidak akan disimpan: " + err.Error())
			s.ScanOptions.SaveToDB = false
		} else {
			defer targetClient.Close()
		}
	}

	// Lakukan scanning
	result, err := s.ExecuteScan(ctx, sourceClient, targetClient, dbFiltered)
	if err != nil {
		s.Logger.Error(config.LogPrefix + " gagal: " + err.Error())
		return err
	}

	// Tampilkan hasil
	s.DisplayScanResult(result)

	// Print success message jika ada
	if config.SuccessMsg != "" {
		ui.PrintSuccess(config.SuccessMsg)
	}

	return nil
}

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

		// Setup session
		sourceClient, dbFiltered, err := s.PrepareScanSession(ctx, "", false)
		if err != nil {
			s.Logger.Errorf("[%s] Gagal setup session: %v", scanID, err)
			return
		}
		defer sourceClient.Close()

		// Koneksi ke target database
		var targetClient *database.Client
		if s.ScanOptions.SaveToDB {
			targetClient, err = s.ConnectToTargetDB(ctx)
			if err != nil {
				s.Logger.Warnf("[%s] Gagal koneksi ke target database: %v", scanID, err)
				s.ScanOptions.SaveToDB = false
			} else {
				defer targetClient.Close()
			}
		}

		// Lakukan scanning
		s.Logger.Infof("[%s] Scanning %d database...", scanID, len(dbFiltered))
		result, err := s.ExecuteScan(ctx, sourceClient, targetClient, dbFiltered)
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
func (s *Service) ExecuteScan(ctx context.Context, sourceClient *database.Client, targetClient *database.Client, dbNames []string) (*ScanResult, error) {
	startTime := time.Now()

	ui.PrintHeader("SCANNING DATABASE")
	ui.PrintInfo(fmt.Sprintf("Total database yang akan di-scan: %d", len(dbNames)))

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
		ui.PrintInfo("Menyimpan hasil scan ke database...")
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

	// Tampilkan hasil jika diminta
	if s.ScanOptions.DisplayResults {
		s.DisplayDetailResults(detailsMap)
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

// DisplayScanResult menampilkan hasil scanning
func (s *Service) DisplayScanResult(result *ScanResult) {
	ui.PrintHeader("HASIL SCANNING")

	data := [][]string{
		{"Total Database", fmt.Sprintf("%d", result.TotalDatabases)},
		{"Berhasil", ui.ColorText(fmt.Sprintf("%d", result.SuccessCount), ui.ColorGreen)},
		{"Gagal", ui.ColorText(fmt.Sprintf("%d", result.FailedCount), ui.ColorRed)},
		{"Durasi", result.Duration},
	}

	headers := []string{"Metrik", "Nilai"}
	ui.FormatTable(headers, data)

	if len(result.Errors) > 0 {
		ui.PrintWarning(fmt.Sprintf("Terdapat %d error saat menyimpan ke database:", len(result.Errors)))
		for _, errMsg := range result.Errors {
			fmt.Printf("  • %s\n", errMsg)
		}
	}
}

// DisplayDetailResults menampilkan detail hasil scanning
func (s *Service) DisplayDetailResults(detailsMap map[string]database.DatabaseDetailInfo) {
	ui.PrintHeader("DETAIL HASIL SCANNING")

	headers := []string{"Database", "Size", "Tables", "Procedures", "Functions", "Views", "Grants", "Status"}
	var rows [][]string

	for _, detail := range detailsMap {
		status := ui.ColorText("✓ OK", ui.ColorGreen)
		if detail.Error != "" {
			status = ui.ColorText("✗ Error", ui.ColorRed)
		}

		rows = append(rows, []string{
			detail.DatabaseName,
			detail.SizeHuman,
			fmt.Sprintf("%d", detail.TableCount),
			fmt.Sprintf("%d", detail.ProcedureCount),
			fmt.Sprintf("%d", detail.FunctionCount),
			fmt.Sprintf("%d", detail.ViewCount),
			fmt.Sprintf("%d", detail.UserGrantCount),
			status,
		})
	}

	ui.FormatTable(headers, rows)
}
