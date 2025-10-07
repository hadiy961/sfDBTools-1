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
			logger.Error(fmt.Sprintf("Validasi konfigurasi gagal: %v", err))
			return
		}
		logger.Debug("Validasi konfigurasi selesai.")
	},
}

func init() {
	// Register dynamic flags for the show command
	flags.AddDBConfigShowFlags(DBConfigValidateCMD)
}
