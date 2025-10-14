// File : cmd/backup_cmd/backup_main.go
// Deskripsi perintah utama 'backup' untuk mengelola backup database
// Author: Hadiyatna Muflihun
// Tanggal: 2024-10-08
// Last Modified: 2024-10-08

package backup_cmd

import (
	"sfDBTools/pkg/globals"

	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/applog"

	"github.com/spf13/cobra"
)

// BackupCMD adalah perintah induk (parent command) untuk semua perintah 'backup'.
var BackupCMD = &cobra.Command{
	Use:   "backup",
	Short: "Mengelola backup database (create, restore, delete, list)",
	Long: `Perintah 'backup' digunakan untuk mengelola proses backup database.
Terdapat beberapa sub-perintah seperti create, restore, delete, dan list.
Gunakan 'backup <sub-command> --help' untuk informasi lebih lanjut tentang masing-masing sub-perintah.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Tambahkan sub-command ke parent command
	BackupCMD.AddCommand(BackupCMDAll)
	BackupCMD.AddCommand(BackupCleanupCmd)
	BackupCMD.AddCommand(BackupCMDDatabase)
}

// GetLogger, GetConfig adalah fungsi helper sederhana untuk modul ini
func GetLogger() applog.Logger {
	return globals.GetLogger()
}

func GetConfig() *appconfig.Config {
	return globals.GetConfig()
}
