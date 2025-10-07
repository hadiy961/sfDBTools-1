package cmd

import (
	"fmt"
	"os"
	"sfDBTools/cmd/dbconfig_cmd"
	"sfDBTools/pkg/globals"

	"github.com/spf13/cobra"
	// Import globals dan sub-command
)

// rootCmd merepresentasikan perintah dasar ketika tidak ada sub-command yang dipanggil
var rootCmd = &cobra.Command{
	Use:   "sfdbtools",
	Short: "SFDBTools: Database Backup and Management Utility",
	Long: `SFDBTools adalah utilitas manajemen dan backup MariaDB/MySQL.
Didesain untuk keandalan dan penggunaan di lingkungan produksi.`,

	// PersistentPreRunE akan dijalankan SEBELUM perintah apapun.
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if globals.Deps == nil || globals.Deps.Config == nil || globals.Deps.Logger == nil {
			return fmt.Errorf("dependency (config/logger) belum diinisialisasi. Error internal.")
		}

		// Log bahwa perintah akan dieksekusi
		globals.GetLogger().Infof("Memulai perintah: %s", cmd.Name())

		return nil
	},

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Silakan jalankan 'sfdbtools --help' untuk melihat perintah yang tersedia.")
	},
}

// Execute adalah fungsi eksekusi utama yang dipanggil dari main.go.
func Execute(deps *globals.Dependencies) {
	// 1. INJEKSI DEPENDENSI
	globals.Deps = deps

	// 2. Eksekusi perintah Cobra
	if err := rootCmd.Execute(); err != nil {
		if globals.Deps != nil && globals.Deps.Logger != nil {
			globals.Deps.Logger.Fatalf("Gagal menjalankan perintah: %v", err)
		} else {
			fmt.Fprintf(os.Stderr, "Gagal menjalankan perintah: %v\n", err)
			os.Exit(1)
		}
	}
}

func init() {
	// Tambahkan sub-command yang sudah dibuat
	// Kita anggap 'versionCmd' ada di cmd/version.go
	rootCmd.AddCommand(versionCmd) // (Perlu diinisialisasi di cmd/version.go)
	rootCmd.AddCommand(dbconfig_cmd.DBConfigMainCMD)
}
