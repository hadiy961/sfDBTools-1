// File : cmd/backup_cmd/backup_cleanup_cmd.go
// Deskripsi : Command untuk cleanup backup manual
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-14
// Last Modified : 2024-10-14

package backup_cmd

import (
	"fmt"
	"sfDBTools/internal/backup"
	"sfDBTools/internal/structs"
	flags "sfDBTools/pkg/flag"
	"sfDBTools/pkg/globals"
	"sfDBTools/pkg/parsing"

	"github.com/spf13/cobra"
)

// BackupCleanupCmd adalah command untuk cleanup backup manual
var BackupCleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Bersihkan backup lama berdasarkan retention policy",
	Long: `Command 'cleanup' membersihkan file backup lama berdasarkan retention policy yang dikonfigurasi.
Backup yang lebih lama dari retention days akan dihapus secara permanen.

Contoh penggunaan:
  sfdbtools backup cleanup --cleanup-days 7 --output /mnt/nfs/backup
  sfdbtools backup cleanup --cleanup --cleanup-days 30`,

	Example: `  # Cleanup backup dengan retention 7 hari
  backup cleanup --cleanup-days 7 --output /path/to/backup

  # Cleanup dengan konfigurasi default
  backup cleanup --cleanup`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validasi dependensi global
		if globals.Deps == nil || globals.Deps.Config == nil || globals.Deps.Logger == nil {
			return fmt.Errorf("dependency (config/logger) belum diinisialisasi")
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		logger := globals.GetLogger()
		cfg := globals.GetConfig()

		logger.Info("Memulai proses cleanup backup manual...")

		// Parse flags khusus cleanup
		cleanupFlags, err := parsing.ParseCleanupFlags(cmd)
		if err != nil {
			logger.Errorf("Gagal mem-parse flags: %v", err)
			return err
		}

		// Debug: Log flags yang diterima
		logger.Debugf("Cleanup enabled: %v", cleanupFlags.Enabled)
		logger.Debugf("Retention days: %d", cleanupFlags.RetentionDays)
		logger.Debugf("Output directory: %s", cleanupFlags.OutputDirectory)
		logger.Debugf("Dry run: %v", cleanupFlags.DryRun)
		logger.Debugf("Pattern: %s", cleanupFlags.Pattern)

		// Pastikan cleanup diaktifkan
		if !cleanupFlags.Enabled {
			logger.Warn("Cleanup tidak diaktifkan, mohon gunakan flag --cleanup untuk mengaktifkan")
			return fmt.Errorf("cleanup harus diaktifkan dengan flag --cleanup")
		}

		// Konversi CleanupFlags ke BackupAllFlags untuk kompatibilitas dengan service
		backupAllFlags := &structs.BackupAllFlags{
			BackupOptions: structs.BackupOptions{
				OutputDirectory: cleanupFlags.OutputDirectory,
				Cleanup: structs.CleanupOptions{
					Enabled:       cleanupFlags.Enabled,
					RetentionDays: cleanupFlags.RetentionDays,
				},
			},
		}

		// Buat service backup
		svc := backup.NewService(logger, cfg, backupAllFlags)

		// Jalankan cleanup berdasarkan mode
		if cleanupFlags.Pattern != "" {
			// Cleanup dengan pattern khusus
			if err := svc.CleanupByPattern(cleanupFlags.Pattern); err != nil {
				logger.Errorf("Cleanup dengan pattern gagal: %v", err)
				return err
			}
		} else {
			// Cleanup normal
			if cleanupFlags.DryRun {
				// Mode dry-run: tampilkan file yang akan dihapus tanpa menghapus
				if err := svc.CleanupDryRun(); err != nil {
					logger.Errorf("Cleanup dry-run gagal: %v", err)
					return err
				}
			} else {
				// Cleanup sebenarnya
				if err := svc.CleanupOldBackups(); err != nil {
					logger.Errorf("Cleanup gagal: %v", err)
					return err
				}
			}
		}

		logger.Info("Cleanup backup manual selesai")
		return nil
	},
}

func init() {
	// Daftarkan flags cleanup yang lebih sederhana
	flags.AddCleanupFlags(BackupCleanupCmd)
}
