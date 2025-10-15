// File : internal/default_value/default_dbscan.go
// Deskripsi : Default values untuk database scan options
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025

package defaultvalue

import (
	"os"
	"sfDBTools/internal/structs"
)

// GetDefaultScanOptions mengembalikan default options untuk database scan
func GetDefaultScanOptions() structs.ScanOptions {
	opts := structs.ScanOptions{}

	// Database Configuration
	opts.DBConfig.FilePath = os.Getenv("SFDB_CONFIG_FILE")

	// Encryption
	opts.Encryption.Key = os.Getenv("SFDB_ENCRYPTION_KEY")

	// Database Selection
	opts.DatabaseList.File = ""
	opts.DatabaseList.UseFile = false

	// Filter Options
	opts.ExcludeSystem = true

	// Target Database (untuk menyimpan hasil scan)
	opts.TargetDB.Host = os.Getenv("SFDB_TARGET_DB_HOST")
	if opts.TargetDB.Host == "" {
		opts.TargetDB.Host = "localhost"
	}

	opts.TargetDB.Port = 3306
	opts.TargetDB.User = os.Getenv("SFDB_TARGET_DB_USER")
	if opts.TargetDB.User == "" {
		opts.TargetDB.User = "root"
	}

	opts.TargetDB.Password = os.Getenv("SFDB_TARGET_DB_PASSWORD")
	opts.TargetDB.Database = os.Getenv("SFDB_TARGET_DB_NAME")
	if opts.TargetDB.Database == "" {
		opts.TargetDB.Database = "sfDBTools"
	}

	// Output Options
	opts.DisplayResults = true
	opts.SaveToDB = false
	opts.Background = false

	return opts
}
