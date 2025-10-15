// File : cmd/backup_cmd/backup_summary_cmd.go
// Deskripsi : Command untuk menampilkan summary backup
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025

package backup_cmd

import (
	"sfDBTools/internal/backup"
	"sfDBTools/internal/structs"
	flags "sfDBTools/pkg/flag"
	"sfDBTools/pkg/globals"
	"sfDBTools/pkg/parsing"

	"github.com/spf13/cobra"
)

// Flags akan di-parse di runtime, tidak perlu global variable

// SummaryCmd adalah command untuk menampilkan summary backup
var SummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Menampilkan summary backup",
	Long: `Command untuk menampilkan summary backup dalam format table.
Summary disimpan dalam direktori base_directory/summaries/ sesuai konfigurasi.

Contoh penggunaan:
  # Menampilkan daftar semua summary
  sfdbtools backup summary

  # Menampilkan summary terbaru  
  sfdbtools backup summary --latest

  # Menampilkan summary berdasarkan backup ID
  sfdbtools backup summary --backup-id=backup_20251015_034246`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ambil dependency yang sudah di-inject
		logger := globals.GetLogger()
		cfg := globals.GetConfig()

		// Parse flags dari command
		backupSummaryFlags, err := parsing.ParseBackupSummaryFlags(cmd)
		if err != nil {
			logger.Errorf("Gagal mem-parse flags: %v", err)
			return err
		}

		// Debug flags
		logger.Debugf("BackupSummaryFlags.BackupID: %s, Latest: %v", backupSummaryFlags.BackupID, backupSummaryFlags.Latest)

		// Inisialisasi service backup
		svc := backup.NewService(logger, cfg, backupSummaryFlags)

		// Tentukan aksi berdasarkan flag
		if backupSummaryFlags.Latest {
			// Tampilkan summary terbaru
			return svc.ShowLatestSummary()
		} else if backupSummaryFlags.BackupID != "" {
			// Tampilkan summary berdasarkan ID
			return svc.ShowSummaryByID(backupSummaryFlags.BackupID)
		} else {
			// Tampilkan daftar semua summary
			return svc.ListAvailableSummaries()
		}
	},
}

func init() {
	// Daftarkan command ke parent
	BackupCMD.AddCommand(SummaryCmd)

	// Daftarkan flags dinamis dengan contoh struct
	dummyFlags := &structs.BackupSummaryFlags{}
	flags.DynamicAddFlags(SummaryCmd, dummyFlags)
}
