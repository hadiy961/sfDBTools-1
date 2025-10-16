// File : internal/structs/structs_dbscan.go
// Deskripsi : Struktur data untuk database scan
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025

package structs

// ScanOptions berisi opsi untuk database scan
type ScanOptions struct {
	// Database Configuration
	DBConfig DBConfigInfo

	// Encryption
	Encryption struct {
		Key string
	}

	// Database Selection
	DatabaseList struct {
		File      string
		Databases []string
		UseFile   bool
	}

	// Filter Options
	ExcludeSystem bool
	IncludeList   []string
	ExcludeList   []string

	// Target Database untuk menyimpan hasil scan
	TargetDB struct {
		Host     string
		Port     int
		User     string
		Password string
		Database string
	}

	// Output Options
	DisplayResults bool
	SaveToDB       bool
	Background     bool // Jalankan scanning di background

	// Internal use only
	Mode string // "all" atau "database"
}
