// File: internal/backup/backup_config.go
// Deskripsi : Fungsi untuk memuat dan menampilkan konfigurasi koneksi database
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-15
// Last Modified : 16 Oktober 2025

package backup

import (
	"sfDBTools/pkg/dbconfig"
)

// CheckAndSelectConfigFile memeriksa dan menampilkan detail file konfigurasi yang dipilih.
// Fungsi ini sekarang menggunakan fungsi generic dari pkg/dbconfig untuk menghindari duplikasi kode.
func (s *Service) CheckAndSelectConfigFile() error {
	err := dbconfig.CheckAndSelectConfigFile(
		&s.BackupOptions.DBConfig,
		s.BackupOptions.Encryption.Key,
		"Pilih file konfigurasi database sumber:",
	)

	// Handle user cancellation khusus untuk backup
	if err == ErrUserCancelled {
		s.Logger.Warn("Proses backup dibatalkan oleh pengguna.")
		return ErrUserCancelled
	}

	return err
}
