package dbconfig

// File: pkg/dbconfig/config_loader.go
// Deskripsi: Fungsi umum untuk memuat dan menampilkan konfigurasi database
// Author: Hadiyatna Muflihun
// Tanggal: 16 Oktober 2025
// Last Modified: 16 Oktober 2025

import (
	"fmt"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/common"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/globals"
	"sfDBTools/pkg/ui"
)

// CheckAndSelectConfigFile memeriksa file konfigurasi yang ada atau memandu pengguna untuk memilihnya.
// Fungsi ini generic dan bisa digunakan oleh berbagai module (backup, dbscan, dll).
//
// Parameters:
//   - configInfo: pointer ke structs.DBConfigInfo yang akan diisi dengan informasi konfigurasi
//   - encryptionKey: kunci enkripsi untuk mendekripsi file konfigurasi
//   - promptMessage: pesan yang ditampilkan saat mode interaktif (optional, default: "Pilih file konfigurasi database:")
//
// Returns:
//   - error: error jika terjadi kesalahan, nil jika berhasil
func CheckAndSelectConfigFile(configInfo *structs.DBConfigInfo, encryptionKey string, promptMessage string) error {
	logger := globals.GetLogger()
	ui.PrintSubHeader("Memeriksa File Konfigurasi")

	if configInfo.FilePath == "" {
		// Jalankan mode interaktif jika tidak ada file yang ditentukan
		ui.PrintWarning("Tidak ada file konfigurasi yang ditentukan. Menjalankan mode interaktif...")

		if promptMessage == "" {
			promptMessage = "Pilih file konfigurasi database:"
		}

		selectedConfig, err := encrypt.SelectExistingDBConfig(promptMessage)
		if err != nil {
			logger.Warn("Proses pemilihan file konfigurasi gagal: " + err.Error())
			return err
		}
		*configInfo = selectedConfig
	} else {
		// Muat konfigurasi dari file yang sudah ditentukan
		if err := LoadAndApplyConfigFromFile(configInfo, encryptionKey); err != nil {
			return err
		}
	}

	DisplayConnectionInfo(*configInfo)
	return nil
}

// LoadAndApplyConfigFromFile memuat dan menerapkan konfigurasi dari file yang ditentukan.
// Fungsi ini adalah helper internal untuk memusatkan logika pemuatan, parsing, dan penerapan konfigurasi.
//
// Parameters:
//   - configInfo: pointer ke structs.DBConfigInfo yang akan diisi dengan informasi dari file
//   - encryptionKey: kunci enkripsi untuk mendekripsi file konfigurasi
//
// Returns:
//   - error: error jika terjadi kesalahan, nil jika berhasil
func LoadAndApplyConfigFromFile(configInfo *structs.DBConfigInfo, encryptionKey string) error {
	logger := globals.GetLogger()

	absPath, name, err := common.ResolveConfigPath(configInfo.FilePath)
	if err != nil {
		return err
	}
	logger.Infof("Menggunakan konfigurasi dari file: %s (%s)", absPath, name)

	// Memuat dan mendekripsi file konfigurasi
	loadedInfo, err := encrypt.LoadAndParseConfig(absPath, encryptionKey)
	if err != nil {
		// Tetap lanjutkan meskipun gagal memuat detail, namun beri peringatan.
		// Ini memungkinkan validasi path file tanpa harus berhasil mendekripsi.
		logger.Warn("Gagal memuat detail konfigurasi untuk validasi: " + err.Error())
	}

	// Terapkan informasi yang berhasil dimuat
	if loadedInfo != nil {
		MergeConnectionInfo(configInfo, loadedInfo)
	}

	// Selalu perbarui path dan nama konfigurasi di state utama
	configInfo.FilePath = absPath
	configInfo.ConfigName = name

	return nil
}

// MergeConnectionInfo menerapkan detail koneksi dari file yang dimuat ke config info.
// Fungsi ini mencegah penimpaan password jika sudah disediakan melalui cara lain (misalnya, flag CLI).
//
// Parameters:
//   - target: pointer ke structs.DBConfigInfo yang akan diupdate
//   - source: structs.DBConfigInfo yang berisi informasi dari file
func MergeConnectionInfo(target *structs.DBConfigInfo, source *structs.DBConfigInfo) {
	conn := &target.ServerDBConnection
	conn.Host = source.ServerDBConnection.Host
	conn.Port = source.ServerDBConnection.Port
	conn.User = source.ServerDBConnection.User

	// Hanya isi password dari file config jika belum ada password yang di-set sebelumnya.
	if conn.Password == "" {
		conn.Password = source.ServerDBConnection.Password
	}
}

// DisplayConnectionInfo menampilkan informasi koneksi database dalam format tabel.
// Fungsi ini bisa digunakan untuk menampilkan info konfigurasi setelah dimuat.
//
// Parameters:
//   - info: structs.DBConfigInfo yang berisi informasi koneksi yang akan ditampilkan
func DisplayConnectionInfo(info structs.DBConfigInfo) {
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
