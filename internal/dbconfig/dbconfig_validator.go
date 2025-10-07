package dbconfig

import (
	"fmt"
	"path/filepath"
	"sfDBTools/pkg/common"
	"sfDBTools/pkg/fs"
)

func (s *Service) ValidateInput() error {
	// Validasi untuk mode pembuatan
	if s.DBConfigInfo.ConfigName == "" {
		return fmt.Errorf("nama konfigurasi tidak boleh kosong")
	}
	// Tambahkan validasi lain sesuai kebutuhan
	return nil
}

func (s *Service) CheckConfigurationNameUnique(mode string) error {
	// Implementasi pengecekan dengan absolute path untuk konsistensi
	// Normalisasi nama yang terlibat untuk memastikan konsistensi
	s.DBConfigInfo.ConfigName = common.TrimConfigSuffix(s.DBConfigInfo.ConfigName)
	switch mode {
	case "create":
		abs := s.filePathInConfigDir(s.DBConfigInfo.ConfigName)
		exists := fs.Exists(abs)
		if exists {
			return fmt.Errorf("nama konfigurasi '%s' sudah ada. Silakan pilih nama lain", s.DBConfigInfo.ConfigName)
		}
		return nil
	case "edit":
		// Untuk mode edit kita punya dua skenario:
		// - Jika user tidak merubah nama (ConfigName == OriginalConfigName): pastikan file lama ada
		// - Jika user merubah nama: pastikan target baru TIDAK ada (agar tidak menimpa)
		// Catatan: Jika --file absolute dipakai, gunakan direktori file tersebut sebagai base.
		original := common.TrimConfigSuffix(s.OriginalConfigName)
		newName := common.TrimConfigSuffix(s.DBConfigInfo.ConfigName)

		// Tentukan baseDir: prioritas ke absolute FilePath (saat edit), fallback ke config dir
		baseDir := s.Config.ConfigDir.DatabaseConfig
		if s.DBConfigInfo.FilePath != "" && filepath.IsAbs(s.DBConfigInfo.FilePath) {
			baseDir = filepath.Dir(s.DBConfigInfo.FilePath)
		}

		if original == "" {
			// Tidak ada nama asli yang direkam -> fallback: cek keberadaan target
			targetAbs := filepath.Join(baseDir, common.EnsureConfigExt(newName))
			if !fs.Exists(targetAbs) {
				return fmt.Errorf("file konfigurasi '%s' tidak ditemukan. Silakan pilih nama lain", newName)
			}
			return nil
		}

		// Jika nama tidak berubah, pastikan file lama tetap ada
		if original == newName {
			origAbs := filepath.Join(baseDir, common.EnsureConfigExt(original))
			if !fs.Exists(origAbs) {
				return fmt.Errorf("file konfigurasi asli '%s' tidak ditemukan", original)
			}
			return nil
		}

		// Nama berubah: pastikan target baru tidak ada
		newAbs := filepath.Join(baseDir, common.EnsureConfigExt(newName))
		if fs.Exists(newAbs) {
			return fmt.Errorf("nama konfigurasi tujuan '%s' sudah ada. Silakan pilih nama lain", newName)
		}
		// juga pastikan file original ada (agar dapat dihapus setelah rename)
		origAbs := filepath.Join(baseDir, common.EnsureConfigExt(original))
		if !fs.Exists(origAbs) {
			return fmt.Errorf("file konfigurasi asli '%s' tidak ditemukan", original)
		}

		return nil
	}
	return nil
}
