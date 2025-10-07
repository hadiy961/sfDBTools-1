// File : cmd/cmd_version.go
// Deskripsi : Sub-command untuk menampilkan versi aplikasi
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package cmd

import (
	"fmt"
	"sfDBTools/pkg/globals"

	"github.com/spf13/cobra"
)

// versionCmd adalah sub-command untuk menampilkan versi aplikasi
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Menampilkan versi aplikasi",
	Run: func(cmd *cobra.Command, args []string) {

		// Akses Config dan Logger melalui package globals
		cfg := globals.GetConfig()
		logger := globals.GetLogger()

		if cfg == nil {
			fmt.Println("Gagal mendapatkan konfigurasi.")
			return
		}

		// Log pesan menggunakan logger yang sudah di-parsing
		logger.Infof("Menampilkan versi untuk %s", cfg.General.AppName)

		fmt.Printf("%s v%s\n", cfg.General.AppName, cfg.General.Version)
		fmt.Printf("Kode Klien: %s\n", cfg.General.ClientCode)
	},
}
