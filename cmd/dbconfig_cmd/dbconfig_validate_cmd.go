// File : cmd/dbconfig_cmd/dbconfig_validate_cmd.go
// Deskripsi : Perintah 'validate' untuk memvalidasi file konfigurasi database yang sudah ada
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

var DBConfigValidateCMD = &cobra.Command{
	Use:   "validate",
	Short: "Memvalidasi file konfigurasi database yang sudah ada",
	Run: func(cmd *cobra.Command, args []string) {
		// Akses logger dan config yang sudah di-inject
		logger := GetLogger()
		cfg := GetConfig()

		logger.Info("Memulai proses generate konfigurasi database...")

		// Parse flags for validate (reuse show flags struct since it's file/encryption related)
		DBConfigShow := &structs.DBConfigShowFlags{}
		if err := parsing.DynamicParseFlags(cmd, DBConfigShow); err != nil {
			logger.Error(fmt.Sprintf("Gagal parse flags: %v", err))
			return
		}

		service := dbconfig.NewService(logger, cfg, DBConfigShow)
		if err := service.ValidateDatabaseConfig(); err != nil {
			return
		}
		logger.Debug("Validasi konfigurasi selesai.")
	},
}

func init() {
	// Register dynamic flags for the show command
	flags.AddDBConfigShowFlags(DBConfigValidateCMD)
}
