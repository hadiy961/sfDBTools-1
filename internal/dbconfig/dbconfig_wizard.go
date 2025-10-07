// File : internal/dbconfig/dbconfig_wizard.go
// Deskripsi : Logika untuk wizard interaktif pembuatan dan pengeditan file konfigurasi database
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package dbconfig

import (
	"fmt"
	"sfDBTools/pkg/common"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
)

// runInteractiveWizard menjalankan wizard dan MENGUBAH state s.DBConfigCreate.
// Tidak perlu mengembalikan data, hanya error.
func (s *Service) runInteractiveWizard(mode string) error {
	s.Logger.Debug("Memulai wizard interaktif untuk mode: " + mode)

	// Gunakan loop for tak terbatas untuk mengulang proses jika diperlukan
	for {
		// Jika mode edit, tampilkan daftar file yang ada dan biarkan user memilih
		if mode == "edit" {
			if err := s.promptSelectExistingConfig(); err != nil {
				return err
			}
			// Setelah memilih file, kita dapat mencoba untuk memuat konten dan
			// mengisi s.DBConfigInfo jika implementasi pembacaan/dekripsi tersedia.
			// Untuk sekarang, cukup lanjutkan ke prompt detail agar user dapat mengubah nilai.
		}
		// 1. Prompt untuk nama konfigurasi
		if err := s.promptDBConfigName(mode); err != nil {
			return err
		}

		// 2. Prompt untuk detail koneksi database
		if err := s.promptDBConfigInfo(); err != nil {
			return err
		}

		// Tampilkan ringkasan konfigurasi
		ConfigName := s.DBConfigInfo.ConfigName
		if ConfigName == "" {
			return fmt.Errorf("nama konfigurasi tidak boleh kosong")
		}

		// Tampilkan detail konfigurasi
		s.DisplayDBConfigDetails()

		//Konfirmasi penyimpanan konfigurasi
		var confirmSave bool
		confirmSave, err := input.AskYesNo("Apakah Anda ingin menyimpan konfigurasi ini?", true)
		if err != nil {
			return common.HandleInputError(err)
		}

		// Jika user mengonfirmasi penyimpanan, keluar dari loop
		// Jika tidak, tanyakan apakah ingin mengulang atau keluar
		if confirmSave {
			break // Keluar dari loop dan lanjutkan proses penyimpanan
		} else {
			var confirmExit bool
			// Tanyakan apakah user ingin mengulang atau keluar
			confirmExit, err = input.AskYesNo("Apakah Anda ingin mengulang proses?", false)
			if err != nil {
				return common.HandleInputError(err)
			}
			if !confirmExit {
				return ErrUserCancelled
			} else {
				ui.PrintWarning("Penyimpanan konfigurasi dibatalkan oleh pengguna. Memulai ulang wizard...")
				continue // Mulai ulang wizard
			}
		}
	}
	ui.PrintSuccess("Konfirmasi diterima. Mempersiapkan enkripsi dan penyimpanan...")

	// 1. Dapatkan password enkripsi dari pengguna (atau env var)
	key, source, err := encrypt.ResolveEncryptionKey(s.DBConfigInfo.EncryptionKey)
	if err != nil {
		return fmt.Errorf("gagal mendapatkan password enkripsi: %w", err)
	}

	s.Logger.WithField("Sumber Kunci", source).Debug("Password enkripsi berhasil didapatkan.")
	s.DBConfigInfo.EncryptionKey = key

	return nil
}
