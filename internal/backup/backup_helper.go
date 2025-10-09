// File : internal/backup/backup_helper.go
// Deskripsi : Helper functions untuk modul backup
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-08
// Last Modified : 2024-10-08
package backup

import (
	"context"
	"errors"
	"fmt"
	"sfDBTools/pkg/common"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/ui"
)

var (
	// ErrGTIDUnsupported dikembalikan bila server tidak mendukung GTID (variabel tidak ada)
	ErrGTIDUnsupported = errors.New("gtid variables unsupported on server")
	// ErrGTIDPermissionDenied dikembalikan bila user tidak punya izin membaca GTID variables
	ErrGTIDPermissionDenied = errors.New("permission denied reading gtid variables")

	// ErrUserCancelled adalah sentinel error untuk menandai pembatalan oleh pengguna.
	ErrUserCancelled = errors.New("user_cancelled")
)

// ResolveConnectionFromConfigFile memuat informasi koneksi database dari file konfigurasi
// yang ditentukan dalam BackupOptions.ConfigFile. Jika file tidak ditentukan,
// fungsi ini tidak melakukan apa-apa.
func (s *Service) ResolveConnectionFromConfigFile() error {

	abs, name, err := common.ResolveConfigPath(s.BackupAll.BackupOptions.DBConfig.FilePath)
	if err != nil {
		return err
	}
	s.Logger.Infof("Menggunakan konfigurasi dari file: %s (%s)", abs, name)
	s.BackupAll.BackupOptions.DBConfig.FilePath = abs

	info, err := encrypt.LoadAndParseConfig(abs, s.BackupAll.BackupOptions.Encryption.Key)
	if err != nil {
		s.Logger.Warn("Gagal memuat isi detail konfigurasi untuk validasi: " + err.Error())
	}
	if info != nil {
		s.BackupAll.BackupOptions.DBConfig.ServerDBConnection.Host = info.ServerDBConnection.Host
		s.BackupAll.BackupOptions.DBConfig.ServerDBConnection.Port = info.ServerDBConnection.Port
		s.BackupAll.BackupOptions.DBConfig.ServerDBConnection.User = info.ServerDBConnection.User
		// Preserve password if flags didn't provide one
		if s.BackupAll.BackupOptions.DBConfig.ServerDBConnection.Password == "" {
			s.BackupAll.BackupOptions.DBConfig.ServerDBConnection.Password = info.ServerDBConnection.Password
		}
	}
	return nil
}

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

// getDatabaseList mendapatkan daftar database dari server, menerapkan filter exclude jika ada.
func (s *Service) getDatabaseList(ctx context.Context, client *database.Client) ([]string, error) {
	var databases []string

	rows, err := client.DB().QueryContext(ctx, "SHOW DATABASES")
	if err != nil {
		return nil, errors.New("gagal mendapatkan daftar database: " + err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			s.Logger.Warn("Gagal membaca nama database: " + err.Error())
			continue
		}

		// Terapkan filter exclude
		if s.shouldExcludeDatabase(dbName) {
			s.Logger.Info("Mengecualikan database: " + dbName)
			continue
		}

		databases = append(databases, dbName)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New("terjadi kesalahan saat membaca daftar database: " + err.Error())
	}

	return databases, nil
}

// shouldExcludeDatabase memeriksa apakah sebuah database harus dikecualikan berdasarkan opsi exclude.
func (s *Service) shouldExcludeDatabase(dbName string) bool {
	// Cek exclude sistem database
	if s.BackupAll.Exclude.SystemsDB {
		systemDBs := map[string]bool{
			"information_schema": true,
			"mysql":              true,
			"performance_schema": true,
			"sys":                true,
			"innodb":             true,
		}
		if systemDBs[dbName] {
			return true
		}
	}

	// Cek exclude berdasarkan daftar yang diberikan user
	for _, exclude := range s.BackupAll.Exclude.Databases {
		if dbName == exclude {
			return true
		}
	}

	return false
}
