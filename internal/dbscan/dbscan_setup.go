// File : internal/dbscan/dbscan_setup.go
// Deskripsi : Fungsi setup untuk database scanning session
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 16 Oktober 2025

package dbscan

import (
	"context"
	"fmt"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
)

// PrepareScanSession mengatur seluruh alur persiapan sebelum proses scanning dimulai.
// Fungsi ini sekarang lebih tangguh dalam menangani resource (koneksi database)
// dengan menggunakan `defer` untuk memastikan koneksi ditutup jika terjadi kegagalan.
func (s *Service) PrepareScanSession(ctx context.Context, headerTitle string, showOptions bool) (client *database.Client, dbFiltered []string, err error) {
	if headerTitle != "" {
		ui.Headers(headerTitle)
		s.Logger.Infof("=== %s ===", headerTitle)
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
		if !success && client != nil {
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
	targetConn := s.getTargetDBConfig()

	client, err := database.InitializeDatabase(targetConn)
	if err != nil {
		return nil, fmt.Errorf("gagal koneksi ke target database: %w", err)
	}

	// Verify koneksi dengan ping
	if err := client.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("gagal verifikasi koneksi: %w", err)
	}

	ui.PrintSuccess(fmt.Sprintf("Koneksi ke target database berhasil: %s@%s:%d/%s",
		targetConn.User, targetConn.Host, targetConn.Port, targetConn.Database))

	return client, nil
}

// getTargetDBConfig memisahkan logika pengambilan konfigurasi dari env vars.
// Ini membuat ConnectToTargetDB lebih fokus pada tugas koneksi.
func (s *Service) getTargetDBConfig() structs.ServerDBConnection {
	// Gunakan nilai dari ScanOptions.TargetDB yang sudah diset dari flags
	conn := s.ScanOptions.TargetDB

	// Override dengan env vars jika nilai belum diset
	if conn.Host == "" {
		conn.Host = database.GetEnvOrDefault("SFDB_TARGET_DB_HOST", "localhost")
	}
	if conn.Port == 0 {
		conn.Port = database.GetEnvOrDefaultInt("SFDB_TARGET_DB_PORT", 3306)
	}
	if conn.User == "" {
		conn.User = database.GetEnvOrDefault("SFDB_TARGET_DB_USER", "root")
	}
	if conn.Password == "" {
		conn.Password = database.GetEnvOrDefault("SFDB_TARGET_DB_PASSWORD", "")
	}
	if conn.Database == "" {
		conn.Database = database.GetEnvOrDefault("SFDB_TARGET_DB_NAME", "sfdbtools")
	}

	return structs.ServerDBConnection{
		Host:     conn.Host,
		Port:     conn.Port,
		User:     conn.User,
		Password: conn.Password,
		Database: conn.Database,
	}
}

// GetFilteredDatabases mengambil dan memfilter daftar database sesuai aturan.
// Menggunakan general database filtering system dari pkg/database
func (s *Service) GetFilteredDatabases(ctx context.Context, client *database.Client) ([]string, *DatabaseFilterStats, error) {

	// Setup filter options from ScanOptions. If a whitelist file is enabled, include it.
	filterOpts := database.FilterOptions{
		ExcludeSystem:    s.ScanOptions.ExcludeSystem,
		ExcludeDatabases: s.ScanOptions.ExcludeList,
		IncludeDatabases: s.ScanOptions.IncludeList,
	}

	// Untuk mode single, gunakan SourceDatabase yang telah ditentukan
	if s.ScanOptions.Mode == "single" && s.ScanOptions.SourceDatabase != "" {
		filterOpts.IncludeDatabases = []string{s.ScanOptions.SourceDatabase}
		// Untuk single database, tidak perlu exclude system databases
		filterOpts.ExcludeSystem = false
	}

	// Untuk mode database, gunakan file list jika tersedia
	if s.ScanOptions.Mode == "database" && s.ScanOptions.DatabaseList.File != "" {
		filterOpts.IncludeFile = s.ScanOptions.DatabaseList.File
	}

	// Execute filtering
	filtered, stats, err := database.FilterDatabases(ctx, client, filterOpts)
	if err != nil {
		return nil, nil, err
	}

	// Convert FilterStats to DatabaseFilterStats (untuk compatibility dengan existing code)
	dbStats := &DatabaseFilterStats{
		TotalFound:     stats.TotalFound,
		ToScan:         stats.TotalIncluded,
		ExcludedSystem: stats.ExcludedSystem,
		ExcludedByList: stats.ExcludedByList,
		ExcludedByFile: stats.ExcludedByFile,
		ExcludedEmpty:  stats.ExcludedEmpty,
	}

	return filtered, dbStats, nil
}
