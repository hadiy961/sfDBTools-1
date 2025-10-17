package dbscan

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
	"syscall"
	"time"
)

// ExecuteScanCommand adalah entry point untuk database scan
func (s *Service) ExecuteScanCommand(config ScanEntryConfig) error {
	ctx := context.Background()
	s.ScanOptions.Mode = config.Mode
	// Jika background mode, spawn sebagai daemon process
	if s.ScanOptions.Background {
		if s.ScanOptions.DBConfig.FilePath == "" {
			return fmt.Errorf("background mode memerlukan file konfigurasi database")
		}

		// Check jika sudah running dalam daemon mode
		if os.Getenv("SFDB_DAEMON_MODE") == "1" {
			// Sudah dalam daemon mode, jalankan actual work
			return s.ExecuteScanInBackground(ctx, config)
		}

		// Spawn new process sebagai daemon
		return s.spawnDaemonProcess(config)
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

// spawnDaemonProcess spawns new process sebagai background daemon
func (s *Service) spawnDaemonProcess(config ScanEntryConfig) error {
	// Get executable path
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("gagal mendapatkan executable path: %w", err)
	}

	// Generate scan ID dan log file
	scanID := fmt.Sprintf("scan_%s", time.Now().Format("20060102_150405"))
	logDir := filepath.Join("logs", "dbscan")

	// Create log directory jika belum ada
	if err := os.MkdirAll(logDir, 0755); err != nil {
		s.Logger.Warnf("Gagal membuat direktori log: %v, menggunakan stdout", err)
		logDir = ""
	}

	// PID file path (fixed name to prevent multiple background runs)
	pidFile := filepath.Join(logDir, "dbscan_background.pid")

	// Check if PID file exists and process is running
	if pidFile != "" {
		if data, err := os.ReadFile(pidFile); err == nil {
			// Parse pid
			var existingPID int
			if _, err := fmt.Sscanf(string(data), "%d", &existingPID); err == nil {
				// Check if process exists
				if err := syscall.Kill(existingPID, 0); err == nil {
					return fmt.Errorf("background process sudah berjalan dengan PID %d (pidfile=%s)", existingPID, pidFile)
				}
				// If syscall.Kill returned ESRCH or other, we proceed and remove stale pidfile
				_ = os.Remove(pidFile)
			} else {
				// Can't parse, remove stale file
				_ = os.Remove(pidFile)
			}
		}
	}

	var logFile string
	if logDir != "" {
		logFile = filepath.Join(logDir, fmt.Sprintf("%s.log", scanID))
	}

	// Prepare command dengan semua arguments
	args := os.Args[1:] // Skip executable name
	cmd := exec.Command(executable, args...)

	// Set environment untuk menandai daemon mode
	cmd.Env = append(os.Environ(), "SFDB_DAEMON_MODE=1")

	// Setup log file untuk stdout/stderr
	if logFile != "" {
		outFile, err := os.Create(logFile)
		if err != nil {
			return fmt.Errorf("gagal membuat log file: %w", err)
		}
		defer outFile.Close()

		cmd.Stdout = outFile
		cmd.Stderr = outFile
	}

	// Start process di background
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("gagal memulai background process: %w", err)
	}

	// Write PID file so we can prevent duplicate background runs
	if pidFile != "" {
		_ = os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644)
	}

	// Print informasi
	ui.PrintHeader("DATABASE SCANNING - BACKGROUND MODE")
	ui.PrintSuccess(fmt.Sprintf("Background process dimulai dengan PID: %d", cmd.Process.Pid))
	ui.PrintInfo(fmt.Sprintf("Scan ID: %s", ui.ColorText(scanID, ui.ColorCyan)))

	if logFile != "" {
		ui.PrintInfo(fmt.Sprintf("Log file: %s", ui.ColorText(logFile, ui.ColorCyan)))
		ui.PrintInfo(fmt.Sprintf("PID file: %s", ui.ColorText(pidFile, ui.ColorCyan)))
		ui.PrintInfo(fmt.Sprintf("Monitor dengan: tail -f %s", logFile))
	} else {
		ui.PrintInfo("Logs akan ditulis ke system logger")
	}

	ui.PrintInfo("Process berjalan di background. Gunakan 'ps aux | grep sfdbtools' untuk check status.")

	// Release process (detach dari parent)
	// Don't wait for it to finish

	return nil
}

// setupScanConnections melakukan setup koneksi source dan target database
// Returns: sourceClient, targetClient, dbFiltered, cleanupFunc, error
func (s *Service) setupScanConnections(ctx context.Context, headerTitle string, showOptions bool) (*database.Client, *database.Client, []string, func(), error) {
	// Jika mode rescan, gunakan PrepareRescanSession
	if s.ScanOptions.Mode == "rescan" {
		sourceClient, targetClient, dbFiltered, err := s.PrepareRescanSession(ctx, headerTitle, showOptions)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		// Force enable SaveToDB untuk rescan karena kita perlu update error_message
		s.ScanOptions.SaveToDB = true

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

	// Setup session (koneksi database source) untuk mode normal
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
