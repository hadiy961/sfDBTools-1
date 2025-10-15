// File : pkg/flag/dbscan_flag.go
// Deskripsi : Flag definitions untuk database scan commands
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025

package flags

import (
	"sfDBTools/internal/structs"

	"github.com/spf13/cobra"
)

// AddDbScanFlags menambahkan flags untuk database scan command
func AddDbScanFlags(cmd *cobra.Command, opts *structs.ScanOptions) {
	// Database Configuration Flags
	cmd.Flags().StringVar(&opts.DBConfig.FilePath, "config-file", opts.DBConfig.FilePath,
		"Path ke file konfigurasi database (encrypted)")
	cmd.Flags().StringVar(&opts.Encryption.Key, "encryption-key", opts.Encryption.Key,
		"Encryption key untuk decrypt config file")

	// Database Selection Flags
	cmd.Flags().StringVar(&opts.DatabaseList.File, "db-list-file", opts.DatabaseList.File,
		"File yang berisi daftar database (satu database per baris)")
	cmd.Flags().StringSliceVar(&opts.IncludeList, "include", opts.IncludeList,
		"Daftar database yang akan di-scan (comma-separated)")
	cmd.Flags().StringSliceVar(&opts.ExcludeList, "exclude", opts.ExcludeList,
		"Daftar database yang akan dikecualikan (comma-separated)")

	// Filter Options Flags
	cmd.Flags().BoolVar(&opts.ExcludeSystem, "exclude-system", opts.ExcludeSystem,
		"Kecualikan system databases (information_schema, mysql, performance_schema, sys)")

	// Target Database Flags
	cmd.Flags().StringVar(&opts.TargetDB.Host, "target-host", opts.TargetDB.Host,
		"Host target database untuk menyimpan hasil scan")
	cmd.Flags().IntVar(&opts.TargetDB.Port, "target-port", opts.TargetDB.Port,
		"Port target database")
	cmd.Flags().StringVar(&opts.TargetDB.User, "target-user", opts.TargetDB.User,
		"User target database")
	cmd.Flags().StringVar(&opts.TargetDB.Password, "target-password", opts.TargetDB.Password,
		"Password target database")
	cmd.Flags().StringVar(&opts.TargetDB.Database, "target-database", opts.TargetDB.Database,
		"Nama database target untuk menyimpan hasil scan")

	// Output Options Flags
	cmd.Flags().BoolVar(&opts.DisplayResults, "display-results", opts.DisplayResults,
		"Tampilkan hasil scan di console")
	cmd.Flags().BoolVar(&opts.SaveToDB, "save-to-db", opts.SaveToDB,
		"Simpan hasil scan ke database (database_details dan database_detail_history)")
	cmd.Flags().BoolVar(&opts.Background, "background", opts.Background,
		"Jalankan scanning di background (async mode)")
}
