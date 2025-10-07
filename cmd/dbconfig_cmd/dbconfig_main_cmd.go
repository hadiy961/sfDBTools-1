// File: cmd/dbconfig_cmd/dbconfig_main.go
// Deskripsi perintah 'main' untuk mengelola konfigurasi database
// Author: Hadiyatna Muflihun
// Tanggal: 2024-10-03
// Last Modified: 2024-10-03

package dbconfig_cmd

import (
	"github.com/spf13/cobra"
	// Import package globals untuk mengakses dependencies
	config "sfDBTools/internal/appconfig"
	logger "sfDBTools/internal/applog"
	"sfDBTools/pkg/globals"
)

// DbConfigCmd adalah perintah induk (parent command) untuk semua perintah 'dbconfig'.
var DBConfigMainCMD = &cobra.Command{
	Use:   "dbconfig",
	Short: "Mengelola konfigurasi database (generate, edit, delete, validate, show)",
	Long: `Perintah 'dbconfig' digunakan untuk mengelola file konfigurasi database.
Terdapat beberapa sub-perintah seperti create, edit, delete, validate, dan show.
Gunakan 'dbconfig <sub-command> --help' untuk informasi lebih lanjut tentang masing-masing sub-perintah.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Tambahkan semua sub-command (Perlu diinisialisasi di file masing-masing)
	DBConfigMainCMD.AddCommand(DBConfigCreateCMD)
	DBConfigMainCMD.AddCommand(DBConfigEditCMD)
	DBConfigMainCMD.AddCommand(DBConfigDeleteCMD)
	DBConfigMainCMD.AddCommand(DBConfigValidateCMD)
	DBConfigMainCMD.AddCommand(DBConfigShowCMD)
}

// GetLogger, GetConfig adalah fungsi helper sederhana untuk modul ini
func GetLogger() logger.Logger {
	return globals.GetLogger()
}

func GetConfig() *config.Config {
	return globals.GetConfig()
}
