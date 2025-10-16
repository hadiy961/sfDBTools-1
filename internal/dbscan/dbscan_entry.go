package dbscan

import (
	"context"
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

	// Setup connections
	sourceClient, targetClient, dbFiltered, cleanup, err := s.setupScanConnections(ctx, config.HeaderTitle, config.ShowOptions)
	if err != nil {
		return err
	}
	defer cleanup()

	// Lakukan scanning (foreground mode dengan UI output)
	result, err := s.ExecuteScan(ctx, sourceClient, targetClient, dbFiltered, false)
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

// setupScanConnections melakukan setup koneksi source dan target database
// Returns: sourceClient, targetClient, dbFiltered, cleanupFunc, error
func (s *Service) setupScanConnections(ctx context.Context, headerTitle string, showOptions bool) (*database.Client, *database.Client, []string, func(), error) {
	// Setup session (koneksi database source)
	sourceClient, dbFiltered, err := s.PrepareScanSession(ctx, headerTitle, showOptions)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Koneksi ke target database untuk menyimpan hasil scan
	var targetClient *database.Client
	if s.ScanOptions.SaveToDB {
		targetClient, err = s.ConnectToTargetDB(ctx)
		if err != nil {
			s.Logger.Warn("Gagal koneksi ke target database, hasil scan tidak akan disimpan: " + err.Error())
			s.ScanOptions.SaveToDB = false
		}
	}

	// Cleanup function untuk close semua connections
	cleanup := func() {
		if sourceClient != nil {
			sourceClient.Close()
		}
		if targetClient != nil {
			targetClient.Close()
		}
	}

	return sourceClient, targetClient, dbFiltered, cleanup, nil
}
