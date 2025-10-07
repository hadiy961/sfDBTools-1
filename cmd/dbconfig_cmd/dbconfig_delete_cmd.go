// File: cmd/dbconfig_cmd/dbconfig_delete.go
// Deskripsi perintah 'delete' untuk menghapus file konfigurasi database yang sudah ada
// Author: Hadiyatna Muflihun
// Tanggal: 2024-10-03
// Last Modified: 2024-10-03

package dbconfig_cmd

import (
	"errors"
	"fmt"

	"sfDBTools/internal/dbconfig"

	"github.com/spf13/cobra"
)

var DBConfigDeleteCMD = &cobra.Command{
	Use:   "delete",
	Short: "Menghapus file konfigurasi database yang sudah ada",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Akses logger dan config yang sudah di-inject
		logger := GetLogger()
		cfg := GetConfig()

		logger.Info("Memulai proses menghapus konfigurasi database...")

		// Buat service dbconfig tanpa perlu state khusus
		service := dbconfig.NewService(logger, cfg, nil)

		// Jalankan proses delete dengan prompt konfirmasi
		if err := service.PromptDeleteConfigs(); err != nil {
			if errors.Is(err, dbconfig.ErrUserCancelled) {
				logger.Warn("Dibatalkan oleh pengguna.")
				return nil
			}
			return fmt.Errorf("penghapusan konfigurasi gagal: %w", err)
		}
		return nil
	},
}
