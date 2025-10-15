// File: cmd/backup_cmd/backup_all.go
// Deskripsi perintah 'all' untuk membuat backup semua database
// Author: Hadiyatna Muflihun
// Tanggal: 2024-10-03
// Last Modified: 2024-10-03

package backup_cmd

import (
	"sfDBTools/internal/backup"
	flags "sfDBTools/pkg/flag"
	"sfDBTools/pkg/parsing"
	"sfDBTools/pkg/ui"

	"github.com/spf13/cobra"
)

var BackupCMDAll = &cobra.Command{
	Use:   "all-databases",
	Short: "Membuat backup semua database",
	Long: `Perintah 'all-databases' memungkinkan pengguna untuk membuat backup dari semua database yang terdaftar.
Pengguna dapat memilih untuk mengisi konfigurasi secara interaktif atau menggunakan flag untuk input non-interaktif.
Gunakan 'backup all-databases --help' untuk informasi lebih lanjut tentang opsi yang tersedia.`,
	Example: `  # Membuat backup semua database secara interaktif
  backup all-databases --config-name local --host localhost --port 3306 --username user --password pass --encryption-key mydb

  # Membuat backup semua database dengan input non-interaktif
  backup all-databases --config-name local --host localhost --port 3306 --username user --password pass --encryption-key mydb --interactive=false
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Akses logger dan config yang sudah di-inject
		logger := GetLogger()
		cfg := GetConfig()
		ui.ClearScreen()
		logger.Info("Memulai proses backup database...")

		// Resolve configuration from flags
		BackupAllFlags, err := parsing.ParseBackupAllFlags(cmd)
		if err != nil {
			logger.Errorf("Gagal mem-parse flags: %v", err)
			return
		}

		// Buat service backup dengan state dari flags
		service := backup.NewService(logger, cfg, BackupAllFlags)

		// validasi mode backup
		if BackupAllFlags.Mode != "single" && BackupAllFlags.Mode != "multi" {
			logger.Errorf("Mode backup tidak valid: %s. Gunakan 'single' atau 'multi'.", BackupAllFlags.Mode)
			return
		}

		// Jalankan proses backup berdasarkan mode yang dipilih
		if BackupAllFlags.Mode == "single" {
			if err := service.BackupAllDatabases(); err != nil {
				logger.Errorf("Backup semua database gagal: %v", err)
				return
			}
		} else {
			if err := service.BackupDatabase(); err != nil {
				logger.Errorf("Backup database gagal: %v", err)
				return
			}
		}

	},
}

func init() {
	// Flags khusus untuk perintah 'all'
	flags.AddBackupAllFlags(BackupCMDAll)
}
