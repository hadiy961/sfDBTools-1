// File : internal/dbconfig/dbconfig_validate.go
// Deskripsi : Logika untuk memvalidasi file konfigurasi database
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package dbconfig

import (
	"context"
	"fmt"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
	"strings"
	"time"
)

// ValidateDatabaseConfig mencoba memvalidasi file konfigurasi yang ada.
// Langkah: baca file terenkripsi, dekripsi dengan encryption key (env atau prompt),
// parse INI untuk mendapatkan host/port/user/password, lalu coba open connection.
func (s *Service) ValidateDatabaseConfig() error {
	ui.Headers("Validate Database Configuration")

	// Pastikan ada flag/file
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
		s.DBConfigInfo.ConfigName = name
		if err := s.loadSnapshotFromPath(abs); err != nil {
			s.Logger.Warn("Gagal memuat isi detail konfigurasi untuk validasi: " + err.Error())
		}
	}

	// Pastikan ada file yang dimuat
	if s.OriginalDBConfigInfo == nil || s.OriginalDBConfigInfo.FilePath == "" {
		return fmt.Errorf("tidak ada file konfigurasi yang dimuat untuk divalidasi")
	}

	// Tampilkan informasi file yang akan divalidasi
	filePath := s.OriginalDBConfigInfo.FilePath
	ui.PrintInfo("Mencoba memvalidasi file: " + filePath)

	host := s.OriginalDBConfigInfo.ServerDBConnection.Host
	port := s.OriginalDBConfigInfo.ServerDBConnection.Port
	user := s.OriginalDBConfigInfo.ServerDBConnection.User
	password := s.OriginalDBConfigInfo.ServerDBConnection.Password

	if host == "" || port == 0 || user == "" {
		ui.PrintError("Informasi koneksi tidak lengkap di file .cnf: host/port/user diperlukan")
		return fmt.Errorf("incomplete connection info")
	}

	// Coba koneksi
	s.Logger.Debug("Attempting to connect to DB with host: " + host + ", port: " + fmt.Sprintf("%d", port) + ", user: " + user)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Buat sebuah fungsi literal (closure) yang membungkus panggilan ke PingMySQL
	pingTask := func() error {
		return database.PingMySQL(ctx, host, port, user, password, 5*time.Second)
	}

	// Definisikan pesan untuk ditampilkan
	message := fmt.Sprintf("Menghubungkan ke %s:%d...", host, port)
	successMessage := "Koneksi Berhasil!"
	// Panggil helper dengan tugas yang sudah didefinisikan
	if err := ui.RunWithSpinner(message, successMessage, pingTask); err != nil {
		return err
	}

	ui.PrintSuccess("Berhasil terhubung ke server database menggunakan konfigurasi di file.")
	return nil
}
