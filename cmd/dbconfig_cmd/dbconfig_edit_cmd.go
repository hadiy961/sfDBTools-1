// File: cmd/dbconfig_cmd/dbconfig_edit.go
// Deskripsi perintah 'edit' untuk mengedit file konfigurasi database yang sudah ada
// Author: Hadiyatna Muflihun
// Tanggal: 2024-10-03
// Last Modified: 2024-10-03

package dbconfig_cmd

import (
	"errors"
	"fmt"
	"sfDBTools/internal/dbconfig"
	flags "sfDBTools/pkg/flag"
	"sfDBTools/pkg/parsing"

	"github.com/spf13/cobra"
)

var DBConfigEditCMD = &cobra.Command{
	Use:   "edit",
	Short: "Mengedit file konfigurasi database yang sudah ada",
	Long:  `Perintah 'edit' memungkinkan pengguna mengedit file konfigurasi database yang sudah ada, baik interaktif maupun non-interaktif.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := GetLogger()
		cfg := GetConfig()

		logger.Info("Memulai proses edit konfigurasi database...")

		// Parse flags khusus edit
		DBConfig, err := parsing.ParseDBConfigEditFlags(cmd)
		if err != nil {
			logger.Error(fmt.Errorf("gagal parse flags: %w", err))
			return
		}

		// Buat service dbconfig dengan state dari flags
		service := dbconfig.NewService(logger, cfg, DBConfig)

		// Jalankan proses edit
		if err := service.EditDatabaseConfig(); err != nil {
			if errors.Is(err, dbconfig.ErrUserCancelled) {
				logger.Warn("Dibatalkan oleh pengguna.")
			}
			logger.Error(fmt.Errorf("edit konfigurasi gagal: %w", err))
			return
		}
	},
}

func init() {
	// Flags khusus untuk perintah 'edit'
	flags.AddDBConfigEditFlags(DBConfigEditCMD)
}
