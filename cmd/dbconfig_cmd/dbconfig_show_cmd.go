// File : cmd/dbconfig_cmd/dbconfig_show_cmd.go
// Deskripsi : Perintah 'show' untuk melihat file konfigurasi database yang sudah ada
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package dbconfig_cmd

import (
	"fmt"
	"sfDBTools/internal/dbconfig"
	"sfDBTools/internal/structs"
	flags "sfDBTools/pkg/flag"
	"sfDBTools/pkg/parsing"

	"github.com/spf13/cobra"
)

var DBConfigShowCMD = &cobra.Command{
	Use:   "show",
	Short: "Melihat file konfigurasi database yang sudah ada",
	Run: func(cmd *cobra.Command, args []string) {
		// Akses logger dan config yang sudah di-inject
		logger := GetLogger()
		cfg := GetConfig()

		logger.Info("Memulai proses melihat konfigurasi database...")

		// Parse flags khusus show
		DBConfigShow := &structs.DBConfigShowFlags{}
		if err := parsing.DynamicParseFlags(cmd, DBConfigShow); err != nil {
			logger.Error(fmt.Sprintf("Gagal parse flags: %v", err))
			return
		}

		// Buat service dbconfig tanpa perlu state khusus
		service := dbconfig.NewService(logger, cfg, DBConfigShow)

		if err := service.ShowDatabaseConfig(); err != nil {
			logger.Error(fmt.Sprintf("Melihat konfigurasi gagal: %v", err))
			return
		}
	},
}

func init() {
	// Register dynamic flags for the show command
	flags.AddDBConfigShowFlags(DBConfigShowCMD)
}
