package backup

import (
	"errors"
	"sfDBTools/pkg/common"
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

// CheckAndSelectConfigFile memeriksa dan menampilkan detail file konfigurasi yang dipilih
func (s *Service) CheckAndSelectConfigFile() error {
	// Check flag configuration file
	ui.PrintSubHeader("Memeriksa File Konfigurasi " + s.DBConfigInfo.FilePath)
	if s.DBConfigInfo.FilePath == "" {
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
		s.BackupOptions.DBConfig = DBConfigInfo
		s.DisplayConnectionInfo(DBConfigInfo)
	} else {
		abs, name, err := common.ResolveConfigPath(s.BackupOptions.DBConfig.FilePath)
		if err != nil {
			return err
		}
		var encryptionKey string
		if s.BackupOptions.Encryption.Key == "" {
			encryptionKey = s.BackupOptions.Encryption.Key
		} else {
			encryptionKey = s.BackupOptions.Encryption.Key
		}

		info, err := encrypt.LoadAndParseConfig(abs, encryptionKey)
		if err != nil {
			s.Logger.Warn("Gagal memuat isi detail konfigurasi untuk validasi: " + err.Error())
		}
		if info != nil {
			s.BackupOptions.DBConfig.ServerDBConnection.Host = info.ServerDBConnection.Host
			s.BackupOptions.DBConfig.ServerDBConnection.Port = info.ServerDBConnection.Port
			s.BackupOptions.DBConfig.ServerDBConnection.User = info.ServerDBConnection.User
			// Preserve password if flags didn't provide one
			if s.BackupOptions.DBConfig.ServerDBConnection.Password == "" {
				s.BackupOptions.DBConfig.ServerDBConnection.Password = info.ServerDBConnection.Password
			}
		}
		s.BackupOptions.DBConfig.FilePath = abs
		s.BackupOptions.DBConfig.ConfigName = name
		s.DisplayConnectionInfo(s.BackupOptions.DBConfig)
	}
	return nil
}
