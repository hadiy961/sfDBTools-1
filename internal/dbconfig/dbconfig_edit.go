// File : internal/dbconfig/dbconfig_edit.go
// Deskripsi : Logika untuk pengeditan file konfigurasi database
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03

package dbconfig

import (
	"fmt"
	"sfDBTools/pkg/fs"
	"sfDBTools/pkg/ui"
	"strings"
)

// EditDatabaseConfig menjalankan logika pengeditan file konfigurasi.
func (s *Service) EditDatabaseConfig() error {
	ui.Headers("Database Configuration Editing")
	if s.DBConfigEdit != nil && s.DBConfigEdit.Interactive {
		s.Logger.Info("Mode interaktif diaktifkan. Memulai wizard konfigurasi...")
		if err := s.runInteractiveWizard("edit"); err != nil {
			if err == ErrUserCancelled {
				s.Logger.Warn("Proses pengeditan konfigurasi dibatalkan oleh pengguna.")
				return ErrUserCancelled
			}
			s.Logger.Warn("Proses pengeditan konfigurasi gagal: " + err.Error())
			return err
		}
	} else {
		s.Logger.Info("Mode non-interaktif. Menggunakan flag yang diberikan.")

		// Non-interactive edit harus memiliki flag --file (tersimpan di OriginalConfigName saat NewService)
		if strings.TrimSpace(s.OriginalConfigName) == "" {
			return fmt.Errorf("flag --file wajib disertakan pada mode non-interaktif (contoh: --file /path/to/name atau --file name)")
		}

		// Resolve path dan nama
		absPath, name, err := s.resolveConfigPath(s.OriginalConfigName)
		if err != nil {
			return err
		}
		if !fs.Exists(absPath) {
			ui.PrintWarning(fmt.Sprintf("File konfigurasi '%s' tidak ditemukan.", absPath))
			return fmt.Errorf("file konfigurasi tidak ditemukan: %s", absPath)
		}

		s.DBConfigInfo.ConfigName = name
		s.DBConfigInfo.FilePath = absPath
		s.OriginalConfigName = name

		s.Logger.Info("File ditemukan. Mencoba memuat konten...")
		if err := s.loadSnapshotFromPath(absPath); err != nil {
			ui.PrintWarning(fmt.Sprintf("Gagal memuat isi file '%s': %v. Lanjut dengan data minimum.", absPath, err))
		}

		// Preserve password if flags didn't provide one
		if s.OriginalDBConfigInfo != nil && s.DBConfigInfo.ServerDBConnection.Password == "" {
			s.DBConfigInfo.ServerDBConnection.Password = s.OriginalDBConfigInfo.ServerDBConnection.Password
		}

		// Tampilkan data dari file yang dimuat ke table
		s.DisplayDBConfigDetails()
	}

	// 2. Lakukan validasi keunikan nama
	err := s.CheckConfigurationNameUnique("edit")
	if err != nil {
		// 3. JIKA TIDAK UNIK: Tampilkan error dan loop akan berulang
		ui.PrintError(err.Error())
		return err
	}

	// 2. Panggil fungsi untuk menyimpan file
	if err := s.SaveDBConfig("Edit"); err != nil {
		ui.PrintError(fmt.Sprintf("Proses penyimpanan file GAGAL: %v", err))
		return err
	}

	// Logika selanjutnya, seperti menyimpan file konfigurasi, bisa ditambahkan di sini
	s.Logger.Info("Wizard interaktif selesai.")

	return nil
}
