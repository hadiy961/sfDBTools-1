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

	"github.com/dustin/go-humanize"
)

var (
	// ErrGTIDUnsupported dikembalikan bila server tidak mendukung GTID (variabel tidak ada)
	ErrGTIDUnsupported = errors.New("gtid variables unsupported on server")
	// ErrGTIDPermissionDenied dikembalikan bila user tidak punya izin membaca GTID variables
	ErrGTIDPermissionDenied = errors.New("permission denied reading gtid variables")

	// ErrUserCancelled adalah sentinel error untuk menandai pembatalan oleh pengguna.
	ErrUserCancelled = errors.New("user_cancelled")

	// ErrNoDatabasesToBackup dikembalikan bila tidak ada database untuk di-backup setelah filtering
	ErrNoDatabasesToBackup = errors.New("tidak ada database untuk di-backup setelah filtering")
)

// DatabaseFilterStats menyimpan statistik hasil filtering database
type DatabaseFilterStats struct {
	TotalFound     int    // Total database yang ditemukan
	ToBackup       int    // Database yang akan di-backup
	ExcludedSystem int    // Database sistem yang dikecualikan
	ExcludedByList int    // Database dikecualikan karena blacklist
	ExcludedByFile int    // Database dikecualikan karena tidak ada di whitelist file
	ExcludedEmpty  int    // Database dengan nama kosong/invalid
	FilterMode     string // Mode filter: "whitelist", "blacklist", atau "system_only"
}

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

// logDatabaseStats menampilkan statistik filtering database
func (s *Service) logDatabaseStats(stats *DatabaseFilterStats) {
	s.Logger.Info("=== Statistik Database Filtering ===")
	s.Logger.Infof("Total database ditemukan: %d", stats.TotalFound)
	s.Logger.Infof("Database untuk backup: %d", stats.ToBackup)

	totalExcluded := stats.ExcludedSystem + stats.ExcludedByList + stats.ExcludedByFile + stats.ExcludedEmpty
	s.Logger.Infof("Total database dikecualikan: %d", totalExcluded)

	if stats.ExcludedSystem > 0 {
		s.Logger.Infof("  - Database sistem dikecualikan: %d", stats.ExcludedSystem)
	}
	if stats.ExcludedByList > 0 {
		s.Logger.Infof("  - Database dikecualikan karena blacklist: %d", stats.ExcludedByList)
	}
	if stats.ExcludedByFile > 0 {
		s.Logger.Infof("  - Database dikecualikan karena tidak ada di whitelist: %d", stats.ExcludedByFile)
	}
	if stats.ExcludedEmpty > 0 {
		s.Logger.Infof("  - Database dengan nama invalid: %d", stats.ExcludedEmpty)
	}

	s.Logger.Infof("Mode filtering: %s", stats.FilterMode)
	s.Logger.Info("=====================================")
}

// CheckAndSelectConfigFile memeriksa dan menampilkan detail file konfigurasi yang dipilih
func (s *Service) CheckAndSelectConfigFile() error {
	// Check flag configuration file
	if s.BackupAll.BackupInfo.FilePath != "" {
		// Jika tidak ada file konfigurasi, tampilkan pilihan interaktif
		ui.PrintWarning("Tidak ada file konfigurasi yang ditentukan. Menjalankan mode interaktif...")
		DBConfigInfo, err := encrypt.SelectExistingDBConfig("Pilih file konfigurasi database sumber:")
		if err != nil {
			if err == ErrUserCancelled {
				s.Logger.Warn("Proses backup dibatalkan oleh pengguna.")
				return ErrUserCancelled
			}
			s.Logger.Warn("Proses pemilihan file konfigurasi gagal: " + err.Error())
			return err
		}
		s.BackupAll.BackupOptions.DBConfig = DBConfigInfo
		s.DisplayConnectionInfo(DBConfigInfo)
	} else {
		abs, name, err := common.ResolveConfigPath(s.BackupAll.BackupOptions.DBConfig.FilePath)
		if err != nil {
			return err
		}
		var encryptionKey string
		if s.BackupAll.BackupOptions.Encryption.Key == "" {
			encryptionKey = s.BackupAll.BackupOptions.Encryption.Key
		} else {
			encryptionKey = s.BackupAll.BackupOptions.Encryption.Key
		}

		info, err := encrypt.LoadAndParseConfig(abs, encryptionKey)
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
		s.BackupAll.BackupOptions.DBConfig.FilePath = abs
		s.BackupAll.BackupOptions.DBConfig.ConfigName = name
		s.DisplayConnectionInfo(s.BackupAll.BackupOptions.DBConfig)
	}
	return nil
}

// CaptureGTIDIfNeeded menangani pengambilan GTID jika opsi diaktifkan
func (s *Service) CaptureGTIDIfNeeded(ctx context.Context, client *database.Client) error {
	if s.BackupAll.CaptureGtid {
		// Cek dukungan GTID
		enabled, pos, err := s.GetGTID(ctx, client)

		if err != nil {
			if errors.Is(err, ErrGTIDUnsupported) {
				s.Logger.Warn("Server tidak mendukung GTID: " + err.Error())
			} else if errors.Is(err, ErrGTIDPermissionDenied) {
				s.Logger.Warn("Tidak memiliki izin untuk membaca variabel GTID: " + err.Error())
			} else {
				s.Logger.Warn("Gagal memeriksa dukungan GTID: " + err.Error())
			}
			s.Logger.Warn("Menonaktifkan opsi capture GTID.")
			s.BackupAll.CaptureGtid = false
		}

		if enabled {
			s.Logger.Info("GTID saat ini: " + pos)
			// Simpan posisi GTID awal untuk referensi
			s.BackupAll.BackupInfo.GTIDCaptured = pos
		} else {
			s.Logger.Warn("GTID tidak diaktifkan pada server ini.")
			s.Logger.Warn("Menonaktifkan opsi capture GTID.")
			s.BackupAll.CaptureGtid = false
		}
	}
	return nil
}

// CheckDatabaseSize memeriksa ukuran database sebelum backup (database/database_count.go)
func (s *Service) CheckDatabaseSize(ctx context.Context, client *database.Client, dbName string) (int64, error) {
	size, err := client.GetDatabaseSize(ctx, dbName)
	if err != nil {
		s.Logger.Warn("Gagal mendapatkan ukuran database " + dbName + ": " + err.Error())
		return 0, err
	}

	//convert ukuran ke human readable pakai external package
	sizeHR := humanize.Bytes(uint64(size))

	// Log ukuran database
	s.Logger.Info("Ukuran database " + dbName + ": " + sizeHR)
	return size, nil
}
