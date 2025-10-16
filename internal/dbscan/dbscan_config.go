// File : internal/dbscan/dbscan_config.go
// Deskripsi : Fungsi untuk memuat dan menampilkan konfigurasi koneksi database
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 16 Oktober 2025

package dbscan

import (
	"sfDBTools/pkg/dbconfig"
)

// CheckAndSelectConfigFile memeriksa file konfigurasi yang ada atau memandu pengguna untuk memilihnya.
// Fungsi ini sekarang menggunakan fungsi generic dari pkg/dbconfig untuk menghindari duplikasi kode.
func (s *Service) CheckAndSelectConfigFile() error {
	return dbconfig.CheckAndSelectConfigFile(
		&s.ScanOptions.DBConfig,
		s.ScanOptions.Encryption.Key,
		"Pilih file konfigurasi database sumber:",
	)
}
