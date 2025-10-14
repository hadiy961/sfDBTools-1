// File : internal/dbconfig/dbconfig_validate.go
// Deskripsi : Logika untuk memvalidasi file konfigurasi database
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package dbconfig

import (
	"context"
	"fmt"
	"sfDBTools/pkg/common"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
	"strings"
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
		abs, name, err := common.ResolveConfigPath(s.DBConfigShow.File)
		if err != nil {
			return err
		}
		// Coba muat snapshot dari path
		s.DBConfigInfo.ConfigName = name
		if err := s.loadSnapshotFromPath(abs); err != nil {
			s.Logger.Warn("Gagal memuat isi detail konfigurasi untuk validasi: " + err.Error())
			return err
		}
	}

	// Pastikan ada file yang dimuat
	if s.OriginalDBConfigInfo == nil || s.OriginalDBConfigInfo.FilePath == "" {
		return fmt.Errorf("tidak ada file konfigurasi yang dimuat untuk divalidasi")
	}

	// Tampilkan informasi file yang akan divalidasi
	filePath := s.OriginalDBConfigInfo.FilePath
	ui.PrintInfo("Mencoba memvalidasi file: " + filePath)

	// Gunakan OriginalDBConfigInfo untuk koneksi karena ini adalah data yang sudah di-load dari file
	connInfo := s.OriginalDBConfigInfo.ServerDBConnection

	// Coba koneksi
	s.Logger.Debug("Mencoba koneksi ke DB dengan host : " + connInfo.Host + ", port: " + fmt.Sprintf("%d", connInfo.Port) + ", user: " + connInfo.User)

	client, err := database.InitializeDatabase(connInfo)
	if err != nil {
		// Jangan gunakan log.Fatalf, kembalikan error yang proper
		return fmt.Errorf("gagal saat inisialisasi koneksi database: %w", err)
	}

	// 3. PENTING: Defer Close() tetap dilakukan di sini, di fungsi pemanggil (main).
	// Ini memastikan koneksi ditutup saat fungsi main selesai.
	defer client.Close()

	ui.PrintInfo("Koneksi database berhasil dibuat dan siap digunakan!")

	if err := client.Ping(context.Background()); err != nil {
		s.Logger.Error("Gagal melakukan ping ke database: " + err.Error())
	} else {
		ui.PrintSuccess("Koneksi database berhasil dan ping sukses!")
	}

	return nil
}
