// File : internal/dbscan/dbscan_rescan.go
// Deskripsi : Fungsi khusus untuk rescan database yang gagal
// Author : Hadiyatna Muflihun
// Tanggal : 16 Oktober 2025
// Last Modified : 16 Oktober 2025

package dbscan

import (
	"context"
	"fmt"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
)

// PrepareRescanSession mengatur persiapan khusus untuk rescan mode
// Rescan mode akan mengambil database yang gagal dari database_details
func (s *Service) PrepareRescanSession(ctx context.Context, headerTitle string, showOptions bool) (sourceClient *database.Client, targetClient *database.Client, dbFiltered []string, err error) {
	if headerTitle != "" {
		ui.Headers(headerTitle)
		s.Logger.Infof("=== %s ===", headerTitle)
	}
	if showOptions {
		s.DisplayScanOptions()
	}

	if err = s.CheckAndSelectConfigFile(); err != nil {
		return nil, nil, nil, fmt.Errorf("gagal memuat konfigurasi database: %w", err)
	}

	// Untuk rescan, kita perlu koneksi ke target database dulu untuk query failed databases
	targetClient, err = s.ConnectToTargetDB(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("gagal koneksi ke target database: %w", err)
	}

	// Gunakan defer pattern untuk cleanup jika terjadi error
	var success bool
	defer func() {
		if !success && targetClient != nil {
			targetClient.Close()
		}
	}()

	// Query database yang gagal
	failedDBNames, err := database.GetFailedDatabaseNames(ctx, targetClient)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("gagal mendapatkan list database yang gagal: %w", err)
	}

	if len(failedDBNames) == 0 {
		ui.PrintInfo("Tidak ada database yang gagal untuk di-rescan")
		return nil, nil, nil, fmt.Errorf("tidak ada database yang gagal untuk di-rescan")
	}

	ui.PrintInfo(fmt.Sprintf("Ditemukan %d database yang gagal di-scan sebelumnya", len(failedDBNames)))

	// Koneksi ke source database untuk scanning
	sourceClient, err = database.InitializeDatabase(s.ScanOptions.DBConfig.ServerDBConnection)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("gagal koneksi ke source database: %w", err)
	}

	defer func() {
		if !success && sourceClient != nil {
			sourceClient.Close()
		}
	}()

	// Display stats untuk rescan
	stats := &DatabaseFilterStats{
		TotalFound:     len(failedDBNames),
		ToScan:         len(failedDBNames),
		ExcludedSystem: 0,
		ExcludedByList: 0,
		ExcludedByFile: 0,
		ExcludedEmpty:  0,
	}
	s.DisplayFilterStats(stats)

	success = true
	return sourceClient, targetClient, failedDBNames, nil
}
