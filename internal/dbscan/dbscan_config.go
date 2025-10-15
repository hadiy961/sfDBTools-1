// File : internal/dbscan/dbscan_config.go
// Deskripsi : Fungsi untuk memuat dan menampilkan konfigurasi koneksi database (mirip backup_config.go)
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025

package dbscan

import (
	"fmt"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/common"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/ui"
)

// CheckAndSelectConfigFile memeriksa file konfigurasi yang ada atau memandu pengguna untuk memilihnya.
// Logika telah disederhanakan untuk menghindari duplikasi kode.
func (s *Service) CheckAndSelectConfigFile() error {
	ui.PrintSubHeader("Memeriksa File Konfigurasi")

	if s.ScanOptions.DBConfig.FilePath == "" {
		// Jalankan mode interaktif jika tidak ada file yang ditentukan
		ui.PrintWarning("Tidak ada file konfigurasi yang ditentukan. Menjalankan mode interaktif...")
		DBConfigInfo, err := encrypt.SelectExistingDBConfig("Pilih file konfigurasi database sumber:")
		if err != nil {
			s.Logger.Warn("Proses pemilihan file konfigurasi gagal: " + err.Error())
			return err
		}
		s.ScanOptions.DBConfig = DBConfigInfo
	} else {
		// Muat konfigurasi dari file yang sudah ditentukan
		if err := s.loadAndApplyConfig(s.ScanOptions.DBConfig.FilePath); err != nil {
			return err
		}
	}

	s.DisplayConnectionInfo(s.ScanOptions.DBConfig)
	return nil
}

// loadAndApplyConfig adalah fungsi helper internal untuk memusatkan logika pemuatan,
// parsing, dan penerapan konfigurasi dari path file untuk menghindari duplikasi.
func (s *Service) loadAndApplyConfig(filePath string) error {
	absPath, name, err := common.ResolveConfigPath(filePath)
	if err != nil {
		return err
	}
	s.Logger.Infof("Menggunakan konfigurasi dari file: %s (%s)", absPath, name)

	// Memuat dan mendekripsi file konfigurasi
	loadedInfo, err := encrypt.LoadAndParseConfig(absPath, s.ScanOptions.Encryption.Key)
	if err != nil {
		// Tetap lanjutkan meskipun gagal memuat detail, namun beri peringatan.
		// Ini memungkinkan validasi path file tanpa harus berhasil mendekripsi.
		s.Logger.Warn("Gagal memuat detail konfigurasi untuk validasi: " + err.Error())
	}

	// Terapkan informasi yang berhasil dimuat
	if loadedInfo != nil {
		s.mergeConnectionInfo(loadedInfo)
	}

	// Selalu perbarui path dan nama konfigurasi di state utama
	s.ScanOptions.DBConfig.FilePath = absPath
	s.ScanOptions.DBConfig.ConfigName = name

	return nil
}

// mergeConnectionInfo menerapkan detail koneksi dari file yang dimuat ke state layanan.
// Ini mencegah penimpaan password jika sudah disediakan melalui cara lain (misalnya, flag CLI).
func (s *Service) mergeConnectionInfo(loadedInfo *structs.DBConfigInfo) {
	conn := &s.ScanOptions.DBConfig.ServerDBConnection
	conn.Host = loadedInfo.ServerDBConnection.Host
	conn.Port = loadedInfo.ServerDBConnection.Port
	conn.User = loadedInfo.ServerDBConnection.User

	// Hanya isi password dari file config jika belum ada password yang di-set sebelumnya.
	if conn.Password == "" {
		conn.Password = loadedInfo.ServerDBConnection.Password
	}
}

// DisplayConnectionInfo menampilkan informasi koneksi database dalam format tabel.
func (s *Service) DisplayConnectionInfo(info structs.DBConfigInfo) {
	ui.PrintSubHeader("Informasi Koneksi Database")

	data := [][]string{
		{"Config Name", info.ConfigName},
		{"Host", info.ServerDBConnection.Host},
		{"Port", fmt.Sprintf("%d", info.ServerDBConnection.Port)},
		{"User", info.ServerDBConnection.User},
		{"File Path", info.FilePath},
	}

	headers := []string{"Parameter", "Value"}
	ui.FormatTable(headers, data)
}
