// File : internal/dbscan/dbscan_structs.go
// Deskripsi : Struktur data untuk database scan
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025

package dbscan

// ScanEntryConfig untuk konfigurasi scan entry point
type ScanEntryConfig struct {
	HeaderTitle string
	ShowOptions bool
	SuccessMsg  string
	LogPrefix   string
}

// ScanResult berisi hasil scanning
type ScanResult struct {
	TotalDatabases int
	SuccessCount   int
	FailedCount    int
	Duration       string
	Errors         []string
}

// DatabaseFilterStats menyimpan statistik hasil filtering database.
type DatabaseFilterStats struct {
	TotalFound     int
	ToScan         int
	ExcludedSystem int
	ExcludedByList int
	ExcludedByFile int // Merepresentasikan database yang tidak ada di include list
	ExcludedEmpty  int
}
