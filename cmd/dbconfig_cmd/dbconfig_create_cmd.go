// File: cmd/dbconfig_cmd/dbconfig_create.go
// Deskripsi perintah 'create' untuk membuat file konfigurasi database baru
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

var DBConfigCreateCMD = &cobra.Command{
	Use:   "create",
	Short: "Membuat file konfigurasi database dari template atau input interaktif",
	Long: `Perintah 'create' memungkinkan pengguna membuat file konfigurasi database baru.
Pengguna dapat memilih untuk mengisi konfigurasi secara interaktif atau menggunakan flag untuk input non-interaktif.
Gunakan 'dbconfig create --help' untuk informasi lebih lanjut tentang opsi yang tersedia.`,
	Example: `  # Membuat konfigurasi database secara interaktif
  dbconfig create

  # Membuat konfigurasi database dengan input non-interaktif
  dbconfig create --config-name local --host localhost --port 3306 --username user --password pass --encryption-key mydb --interactive=false
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Akses logger dan config yang sudah di-inject
		logger := GetLogger()
		cfg := GetConfig()

		logger.Info("Memulai proses generate konfigurasi database...")

		// Resolve configuration from flags
		DBConfig, err := parsing.ParseDBConfigCreateFlags(cmd)
		if err != nil {
			return fmt.Errorf("gagal parse flags: %w", err)
		}

		// Buat service dbconfig dengan state dari flags
		service := dbconfig.NewService(logger, cfg, DBConfig)

		// Jalankan proses create
		if err := service.CreateDatabaseConfig(); err != nil {
			if errors.Is(err, dbconfig.ErrUserCancelled) {
				return nil
			}
			return fmt.Errorf("create konfigurasi gagal: %w", err)
		}

		return nil
	},
}

func init() {
	// Flags khusus untuk perintah 'create'
	flags.AddDBConfigCreateFlags(DBConfigCreateCMD)
}
