// File : internal/dbconfig/dbconfig_save.go
// Deskripsi : Logika untuk menyimpan file konfigurasi database
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package dbconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/pkg/common"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/fs"
	"sfDBTools/pkg/ui"
)

func (s *Service) SaveDBConfig(mode string) error {
	s.Logger.Info("Memulai proses penyimpanan file konfigurasi...")
	// 1. Tentukan base directory tujuan penyimpanan.
	//    - Untuk mode Edit dengan --file absolute: simpan di direktori file asli (in-place)
	//    - Selain itu: gunakan direktori konfigurasi aplikasi.
	var baseDir string
	var originalAbsPath string
	if mode == "Edit" && s.DBConfigInfo.FilePath != "" && filepath.IsAbs(s.DBConfigInfo.FilePath) {
		originalAbsPath = s.DBConfigInfo.FilePath
		baseDir = filepath.Dir(s.DBConfigInfo.FilePath)
	} else {
		baseDir = s.Config.ConfigDir.DatabaseConfig
	}

	// 2. Pastikan base directory ada, jika tidak buat (berlaku untuk config dir maupun absolut)
	dir, err := fs.CheckDirExists(baseDir)
	if err != nil {
		return fmt.Errorf("gagal memastikan direktori konfigurasi ada: %v", err)
	}
	s.Logger.Debug("Direktori konfigurasi ada: " + baseDir)
	if !dir {
		s.Logger.Info("Direktori konfigurasi tidak ada. Mencoba membuat: " + baseDir)
		if err := fs.CreateDirIfNotExist(baseDir); err != nil {
			return fmt.Errorf("gagal membuat direktori konfigurasi: %v", err)
		}
		s.Logger.Info("Direktori konfigurasi berhasil dibuat: " + baseDir)
	}

	// 3. Ubah konfigurasi ke format INI
	iniContent := s.formatConfigToINI()

	// 4. Resolusi kunci enkripsi dan enkripsi konten
	key, _, err := encrypt.ResolveEncryptionKey(s.DBConfigInfo.EncryptionKey)
	if err != nil {
		return fmt.Errorf("kunci enkripsi tidak tersedia: %w", err)
	}
	// Enkripsi konten INI
	encryptedContent, err := encrypt.EncryptAES([]byte(iniContent), []byte(key))
	if err != nil {
		return fmt.Errorf("gagal mengenkripsi konten konfigurasi: %v", err)
	}

	// 5. Tentukan nama dan path file tujuan (normalisasi nama agar tidak double suffix)
	// Normalisasi nama simpanan: trim suffix jika ada lalu pastikan .cnf.enc
	s.DBConfigInfo.ConfigName = common.TrimConfigSuffix(s.DBConfigInfo.ConfigName)
	newFileName := buildFileName(s.DBConfigInfo.ConfigName)
	newFilePath := filepath.Join(baseDir, newFileName)
	s.Logger.Debug("File akan disimpan di: " + newFilePath)

	// If mode is Edit and original name exists and different -> perform rename flow
	if mode == "Edit" && s.OriginalConfigName != "" && s.OriginalConfigName != s.DBConfigInfo.ConfigName {
		// 1) Write new file
		if err := fs.WriteFile(newFilePath, encryptedContent); err != nil {
			return fmt.Errorf("gagal menyimpan file konfigurasi baru: %v", err)
		}
		// 2) Delete old file
		oldFilePath := originalAbsPath
		if oldFilePath == "" && s.OriginalDBConfigInfo != nil && s.OriginalDBConfigInfo.FilePath != "" {
			oldFilePath = s.OriginalDBConfigInfo.FilePath
		}
		if oldFilePath == "" {
			// fallback ke baseDir + nama lama jika metadata tidak tersedia
			oldFilePath = filepath.Join(baseDir, buildFileName(s.OriginalConfigName))
		}
		if err := os.Remove(oldFilePath); err != nil {
			// Jika gagal menghapus, laporkan tapi lanjutkan (user mungkin akan membersihkan manual)
			ui.PrintWarning(fmt.Sprintf("Berhasil menyimpan '%s' tetapi gagal menghapus file lama '%s': %v", newFileName, oldFilePath, err))
		}

		ui.PrintSuccess(fmt.Sprintf("File konfigurasi berhasil disimpan sebagai '%s' (rename dari '%s').", newFileName, buildFileName(s.OriginalConfigName)))
		ui.PrintInfo("File konfigurasi tersimpan di : " + newFilePath)

		return nil
	}

	// Default: write/overwrite target file (create or update without removing anything)
	if err := fs.WriteFile(newFilePath, encryptedContent); err != nil {
		return fmt.Errorf("gagal menyimpan file konfigurasi: %v", err)
	}

	ui.PrintSuccess(fmt.Sprintf("File konfigurasi '%s' berhasil disimpan dengan aman.", newFileName))
	ui.PrintInfo("File konfigurasi tersimpan di : " + newFilePath)

	return nil
}

// formatConfigToINI mengubah struct DBConfigInfo menjadi format string INI.
func (s *Service) formatConfigToINI() string {
	// [client] adalah header standar untuk file my.cnf
	// Ini memastikan kompatibilitas dengan banyak tools command-line MySQL/MariaDB
	content := `[client]
host=%s
port=%d
user=%s
password=%s
`
	return fmt.Sprintf(content, s.DBConfigInfo.ServerDBConnection.Host, s.DBConfigInfo.ServerDBConnection.Port, s.DBConfigInfo.ServerDBConnection.User, s.DBConfigInfo.ServerDBConnection.Password)
}
