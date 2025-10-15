package backup

import (
	"sfDBTools/pkg/common"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/ui"
)

// ResolveConnectionFromConfigFile memuat informasi koneksi database dari file konfigurasi
// yang ditentukan dalam BackupOptions.ConfigFile. Jika file tidak ditentukan,
// fungsi ini tidak melakukan apa-apa.
func (s *Service) ResolveConnectionFromConfigFile() error {
	abs, name, err := common.ResolveConfigPath(s.BackupOptions.DBConfig.FilePath)
	if err != nil {
		return err
	}
	s.Logger.Infof("Menggunakan konfigurasi dari file: %s (%s)", abs, name)
	s.BackupOptions.DBConfig.FilePath = abs

	info, err := encrypt.LoadAndParseConfig(abs, s.BackupOptions.Encryption.Key)
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
