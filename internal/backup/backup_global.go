package backup

import (
	"context"
	"fmt"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
	"strings"
)

// KembalikanMaxStatementsTime mengembalikan nilai max_statements_time saat ini dari sesi database.
func (s *Service) KembalikanMaxStatementsTime(ctx context.Context, client *database.Client, original float64) {
	ui.PrintSubHeader("Mengembalikan nilai max_statement_time")
	if err := client.SetMaxStatementsTime(ctx, original); err != nil {
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
func (s *Service) buildMysqldumpArgs(baseDumpArgs string, dbFiltered []string) []string {
	var args []string

	// Tambahkan kredensial database
	dbConn := s.BackupAll.BackupOptions.DBConfig.ServerDBConnection

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

	// Tambahkan database yang sudah difilter
	totalDB := s.FilterInfo.TotalDatabases
	excludedDB := s.FilterInfo.ExcludedDatabases
	toBackupDB := s.FilterInfo.IncludedDatabases

	// Tentukan apakah menggunakan --all-databases atau --databases
	if totalDB == toBackupDB && excludedDB == 0 {
		args = append(args, "--all-databases")
		s.Logger.Infof("Menggunakan --all-databases untuk backup semua database (%d database)", totalDB)
	} else {
		args = append(args, "--databases")
		args = append(args, dbFiltered...)
		s.Logger.Infof("Menggunakan --databases dengan %d database yang dipilih", toBackupDB)
	}

	return args
}
