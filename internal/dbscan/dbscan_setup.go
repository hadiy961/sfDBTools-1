// File : internal/dbscan/dbscan_setup.go
// Deskripsi : Fungsi setup untuk database scanning session
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025

package dbscan

import (
	"context"
	"fmt"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
	"strings"
)

// systemDatabases didefinisikan sebagai package-level variable untuk efisiensi
// dan kemudahan pengelolaan. Menggunakan map untuk lookup O(1).
var systemDatabases = map[string]struct{}{
	"information_schema": {},
	"mysql":              {},
	"performance_schema": {},
	"sys":                {},
}

// PrepareScanSession mengatur seluruh alur persiapan sebelum proses scanning dimulai.
// Fungsi ini sekarang lebih tangguh dalam menangani resource (koneksi database)
// dengan menggunakan `defer` untuk memastikan koneksi ditutup jika terjadi kegagalan.
func (s *Service) PrepareScanSession(ctx context.Context, headerTitle string, showOptions bool) (client *database.Client, dbFiltered []string, err error) {
	if headerTitle != "" {
		ui.PrintHeader(headerTitle)
	}
	if showOptions {
		s.DisplayScanOptions()
	}

	if err = s.CheckAndSelectConfigFile(); err != nil {
		return nil, nil, fmt.Errorf("gagal memuat konfigurasi database: %w", err)
	}

	client, err = database.InitializeDatabase(s.ScanOptions.DBConfig.ServerDBConnection)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal koneksi ke database: %w", err)
	}

	// Gunakan pola `defer` dengan flag untuk memastikan `client.Close()` hanya dipanggil saat terjadi error.
	// Jika fungsi berhasil, client akan dikembalikan dalam keadaan terbuka.
	var success bool
	defer func() {
		if !success {
			client.Close()
		}
	}()

	var stats *DatabaseFilterStats
	dbFiltered, stats, err = s.GetFilteredDatabases(ctx, client)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal mendapatkan daftar database: %w", err)
	}

	s.DisplayFilterStats(stats)

	if len(dbFiltered) == 0 {
		return nil, nil, fmt.Errorf("tidak ada database untuk di-scan setelah filtering")
	}

	success = true // Tandai sebagai sukses agar koneksi tidak ditutup oleh defer.
	return client, dbFiltered, nil
}

// ConnectToTargetDB membuat koneksi ke database pusat untuk menyimpan hasil.
// Logika untuk mendapatkan konfigurasi dipisahkan untuk kejelasan.
func (s *Service) ConnectToTargetDB(ctx context.Context) (*database.Client, error) {
	ui.PrintSubHeader("Koneksi ke Database Center")
	targetConn := s.getTargetDBConfig()

	client, err := database.InitializeDatabase(targetConn)
	if err != nil {
		return nil, fmt.Errorf("gagal koneksi ke target database: %w", err)
	}

	// Uji koneksi dengan memilih database target.
	if targetConn.Database != "" {
		if err := client.Ping(ctx); err != nil {
			client.Close()
			return nil, err
		}
	}

	ui.PrintSuccess(fmt.Sprintf("Koneksi ke target database berhasil: %s@%s:%d/%s",
		targetConn.User, targetConn.Host, targetConn.Port, targetConn.Database))

	return client, nil
}

// getTargetDBConfig memisahkan logika pengambilan konfigurasi dari env vars.
// Ini membuat ConnectToTargetDB lebih fokus pada tugas koneksi.
func (s *Service) getTargetDBConfig() structs.ServerDBConnection {
	// Jika konfigurasi target DB sudah ada di ScanOptions, gunakan itu.
	// Jika tidak, fallback ke environment variables.

	return structs.ServerDBConnection{
		Host:     database.GetEnvOrDefault("SFDB_DB_HOST", "localhost"),
		Port:     database.GetEnvOrDefaultInt("SFDB_DB_PORT", 3306),
		User:     database.GetEnvOrDefault("SFDB_DB_USER", "root"),
		Password: database.GetEnvOrDefault("SFDB_DB_PASSWORD", ""),
		Database: database.GetEnvOrDefault("SFDB_DB_NAME", ""),
	}
}

// GetFilteredDatabases mengambil dan memfilter daftar database sesuai aturan.
// Logika filter yang kompleks dipecah menjadi fungsi-fungsi helper yang lebih kecil.
func (s *Service) GetFilteredDatabases(ctx context.Context, client *database.Client) ([]string, *DatabaseFilterStats, error) {
	allDatabases, err := client.GetDatabaseList(ctx, client)
	if err != nil {
		return nil, nil, fmt.Errorf("tidak dapat mengambil daftar database dari server: %w", err)
	}

	stats := &DatabaseFilterStats{TotalFound: len(allDatabases)}
	filtered := make([]string, 0, len(allDatabases)) // Pra-alokasi slice untuk performa

	for _, dbName := range allDatabases {
		if s.shouldExclude(dbName, stats) {
			continue // Lanjut ke database berikutnya jika harus dikecualikan
		}
		filtered = append(filtered, dbName)
	}

	stats.ToScan = len(filtered)
	return filtered, stats, nil
}

// shouldExclude adalah fungsi helper yang mengoordinasikan semua aturan pengecualian.
func (s *Service) shouldExclude(dbName string, stats *DatabaseFilterStats) bool {
	if dbName == "" {
		stats.ExcludedEmpty++
		return true
	}
	if s.ScanOptions.ExcludeSystem && isSystemDatabase(dbName) {
		stats.ExcludedSystem++
		return true
	}
	// Periksa exclude list terlebih dahulu
	if len(s.ScanOptions.ExcludeList) > 0 && contains(s.ScanOptions.ExcludeList, dbName) {
		stats.ExcludedByList++
		return true
	}
	// Jika include list ditentukan, hanya izinkan yang ada di dalamnya
	if len(s.ScanOptions.IncludeList) > 0 && !contains(s.ScanOptions.IncludeList, dbName) {
		stats.ExcludedByFile++ // Nama stat ini mungkin perlu disesuaikan jika sumbernya bukan file
		return true
	}
	return false
}

// isSystemDatabase memeriksa apakah nama database termasuk database sistem.
// Pemeriksaan menjadi sangat cepat dengan menggunakan map.
func isSystemDatabase(dbName string) bool {
	_, exists := systemDatabases[strings.ToLower(dbName)]
	return exists
}

// contains adalah fungsi generik untuk memeriksa keberadaan item dalam slice.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
