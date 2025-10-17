package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
	"strings"
)

// KembalikanMaxStatementsTime mengembalikan nilai max_statements_time saat ini dari sesi database.
func (s *Service) KembalikanMaxStatementsTime(ctx context.Context, original float64) {
	ui.PrintSubHeader("Mengembalikan nilai max_statement_time")
	if err := s.Client.SetMaxStatementsTime(ctx, original); err != nil {
		s.Logger.Warn("Gagal mengembalikan nilai max_statement_time: " + err.Error())
	} else {
		s.Logger.Info("Nilai max_statement_time berhasil dikembalikan ke " + fmt.Sprintf("%f", original))
	}
}

// AturMaxStatementsTimeJikaPerlu mengatur max_statements_time jika opsi diatur.
// Mengembalikan nilai asli untuk dikembalikan nanti.
func (s *Service) AturMaxStatementsTime(ctx context.Context, client *database.Client) (float64, error) {
	// Set dan cek max_statement_time untuk sesi ini
	ui.PrintSubHeader("Setting max_statement_time")
	// Cek nilai awal
	originalMaxStatementsTime, err := client.GetMaxStatementsTime(ctx)
	if err != nil {
		s.Logger.Warn("Gagal mendapatkan timeout statement awal: " + err.Error())
	} else {
		s.Logger.Info(fmt.Sprintf("Original max_statement_time: %f detik", originalMaxStatementsTime))
	}

	// Set nilai baru
	s.Logger.Info("Mengatur max_statement_time ke 0 detik untuk sesi ini...")
	// 0 berarti tidak ada batasan waktu
	if err := client.SetMaxStatementsTime(ctx, 0); err != nil {
		s.Logger.Warn("Gagal mengatur timeout statement: " + err.Error())
	}

	// Verifikasi nilai baru
	currentTimeout, err := client.GetMaxStatementsTime(ctx)
	if err != nil {
		s.Logger.Warn("Gagal mendapatkan timeout statement saat ini: " + err.Error())
	} else {
		s.Logger.Debug(fmt.Sprintf("Current max_statement_time: %f detik", currentTimeout))
	}

	return originalMaxStatementsTime, nil
}

// buildMysqldumpArgs membangun argumen mysqldump dengan kredensial database
// Parameter singleDB: jika tidak kosong, akan backup database tunggal tersebut
// Parameter dbFiltered: list database untuk backup multiple (diabaikan jika singleDB diisi)
func (s *Service) buildMysqldumpArgs(baseDumpArgs string, dbFiltered []string, singleDB string) []string {
	var args []string

	// Tambahkan kredensial database
	dbConn := s.BackupOptions.DBConfig.ServerDBConnection

	// Host
	if dbConn.Host != "" {
		args = append(args, "--host="+dbConn.Host)
	}

	// Port
	if dbConn.Port != 0 {
		args = append(args, fmt.Sprintf("--port=%d", dbConn.Port))
	}

	// User
	if dbConn.User != "" {
		args = append(args, "--user="+dbConn.User)
	}

	// Password
	if dbConn.Password != "" {
		args = append(args, "--password="+dbConn.Password)
	}

	// Tambahkan argumen mysqldump dari konfigurasi
	if baseDumpArgs != "" {
		baseArgs := strings.Fields(baseDumpArgs)
		args = append(args, baseArgs...)
	}

	// jika exclude-data diaktifkan, tambahkan --no-data
	if s.BackupOptions.Exclude.Data {
		args = append(args, "--no-data")
	}

	// Mode single database
	if singleDB != "" {
		args = append(args, "--databases")
		args = append(args, singleDB)
		return args
	}

	// Mode multiple databases
	totalDB := s.FilterInfo.TotalDatabases
	excludedDB := s.FilterInfo.ExcludedDatabases
	toBackupDB := s.FilterInfo.IncludedDatabases

	// Tentukan apakah menggunakan --all-databases atau --databases
	if totalDB == toBackupDB && excludedDB == 0 {
		args = append(args, "--all-databases")
	} else {
		args = append(args, "--databases")
		args = append(args, dbFiltered...)
	}

	return args
}

// GetAndFilterDatabases mendapatkan dan memfilter database berdasarkan opsi exclude
func (s *Service) GetAndFilterDatabases(ctx context.Context, client *database.Client) ([]string, error) {
	ui.PrintSubHeader("Mendapatkan dan Memfilter Database")
	var DBList string
	// Setup filter options
	if s.BackupOptions.UseDBList && s.BackupOptions.DBList != "" {
		s.Logger.Info("Menggunakan file database list untuk memfilter database...")
		DBList = s.BackupOptions.DBList
	} else {
		s.Logger.Info("Tidak menggunakan file database list untuk memfilter database...")
		DBList = ""
	}

	filterOpts := database.FilterOptions{
		ExcludeSystem:    s.BackupOptions.Exclude.SystemsDB,
		ExcludeDatabases: s.BackupOptions.Exclude.Databases,
		IncludeFile:      DBList, // Whitelist file (jika ada)
	}

	// Execute filtering
	s.Logger.Info("Mendapatkan daftar database dari server...")
	validDatabases, stats, err := database.FilterDatabases(ctx, client, filterOpts)
	if err != nil {
		s.Logger.Error("Gagal memfilter database: " + err.Error())
		return nil, err
	}

	// Log statistics
	s.Logger.Infof("Ditemukan %d database di server.", stats.TotalFound)
	s.Logger.Infof("Total database: %d, Untuk backup: %d, Dikecualikan: %d",
		stats.TotalFound, stats.TotalIncluded, stats.TotalExcluded)

	if stats.TotalExcluded == 0 {
		s.Logger.Info("Tidak ada database yang dikecualikan; semua database akan di-backup.")
	}

	// Simpan statistik filtering ke struct (mapping dari FilterStats ke FilterInfo)
	s.FilterInfo = &structs.FilterInfo{
		TotalDatabases:    stats.TotalFound,
		ExcludedDatabases: stats.TotalExcluded,
		IncludedDatabases: stats.TotalIncluded,
		SystemDatabases:   stats.ExcludedSystem,
	}

	return validDatabases, nil
}
