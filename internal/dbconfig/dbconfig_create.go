package dbconfig

import (
	"fmt"
	"sfDBTools/pkg/ui"
)

// CreateDatabaseConfig menjalankan logika pembuatan file konfigurasi.
func (s *Service) CreateDatabaseConfig() error {
	ui.Headers("Database Configuration Creation")
	if s.DBConfigCreate != nil && s.DBConfigCreate.Interactive {
		s.Logger.Info("Mode interaktif diaktifkan. Memulai wizard konfigurasi...")
		if err := s.runInteractiveWizard("create"); err != nil {
			if err == ErrUserCancelled {
				s.Logger.Warn("Proses pembuatan konfigurasi dibatalkan oleh pengguna.")
				return ErrUserCancelled
			}
			s.Logger.Warn("Proses pembuatan konfigurasi gagal: " + err.Error())
			return err
		}
	} else {
		s.Logger.Info("Mode non-interaktif. Menggunakan flag yang diberikan.")

		s.DisplayDBConfigDetails()
	}

	// 2. Lakukan validasi keunikan nama
	err := s.CheckConfigurationNameUnique("create")
	if err != nil {
		// 3. JIKA TIDAK UNIK: Tampilkan error dan loop akan berulang
		ui.PrintError(err.Error())
		return err
	}

	// 2. Panggil fungsi untuk menyimpan file
	if err := s.SaveDBConfig("Create"); err != nil {
		ui.PrintError(fmt.Sprintf("Proses penyimpanan file GAGAL: %v", err))
		return err
	}

	// Logika selanjutnya, seperti menyimpan file konfigurasi, bisa ditambahkan di sini
	s.Logger.Info("Wizard interaktif selesai.")

	return nil
}
