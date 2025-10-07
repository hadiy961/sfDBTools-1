// File : internal/dbconfig/dbconfig_show.go
// Deskripsi : Logika untuk menampilkan detail file konfigurasi database
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

// ShowDatabaseConfig menampilkan detail konfigurasi database yang ada.
func (s *Service) ShowDatabaseConfig() error {
	ui.Headers("Display Database Configuration")

	// 1. Pastikan flag --file diisi

	// Guard against nil or empty flags
	if s.DBConfigShow == nil || strings.TrimSpace(s.DBConfigShow.File) == "" {
		if err := s.promptSelectExistingConfig(); err != nil {
			return err
		}
	} else {
		// Jika user memberikan --file, resolve path dan muat snapshot
		abs, name, err := s.resolveConfigPath(s.DBConfigShow.File)
		if err != nil {
			return err
		}
		// Coba muat snapshot dari path
		if !fs.Exists(abs) {
			ui.PrintWarning(fmt.Sprintf("File konfigurasi '%s' tidak ditemukan.", abs))
			return fmt.Errorf("file konfigurasi tidak ditemukan: %s", abs)
		}
		// Muat snapshot dari path
		s.DBConfigInfo.ConfigName = name
		if err := s.loadSnapshotFromPath(abs); err != nil {
			// Tetap tampilkan metadata minimum jika ada, dan informasikan errornya
			s.Logger.Warn("Gagal memuat isi detail konfigurasi: " + err.Error())
		}
	}

	// Pastikan ada file yang dimuat
	if s.OriginalDBConfigInfo == nil || s.OriginalDBConfigInfo.FilePath == "" {
		return fmt.Errorf("tidak ada snapshot konfigurasi untuk ditampilkan")
	}

	// 2. Pastikan file ada
	s.Logger.Debug("Mengecek file konfigurasi: " + s.OriginalDBConfigInfo.FilePath)
	if !fs.Exists(s.OriginalDBConfigInfo.FilePath) {
		ui.PrintWarning(fmt.Sprintf("File konfigurasi '%s' tidak ditemukan.", s.OriginalDBConfigInfo.FilePath))
		return fmt.Errorf("file konfigurasi tidak ditemukan: %s", s.OriginalDBConfigInfo.FilePath)
	}

	// 3. Tampilkan detail
	s.DBConfigInfo.FilePath = s.OriginalDBConfigInfo.FilePath
	s.Logger.Info("File ditemukan. Menampilkan konten...")
	s.DisplayDBConfigDetails()

	return nil
}
