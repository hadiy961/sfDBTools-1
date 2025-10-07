// File : internal/dbconfig/dbconfig_display.go
// Deskripsi : Logika untuk menampilkan detail file konfigurasi database
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03

package dbconfig

import (
	"fmt"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"

	"github.com/AlecAivazis/survey/v2"
)

// Menampilkan detail konfigurasi database yang telah dibuat
func (s *Service) DisplayDBConfigDetails() {
	// Jika mode SHOW -> tampilkan detail penuh dari snapshot yang dimuat
	if s.DBConfigShow != nil && s.OriginalDBConfigInfo != nil {
		title := s.OriginalDBConfigInfo.ConfigName
		if title == "" {
			title = s.DBConfigInfo.ConfigName
		}
		ui.PrintSubHeader("Menampilkan Konfigurasi: " + title)
		s.printShowDetails()
		return
	}

	// Jika ada snapshot original -> ini alur EDIT: tampilkan hanya ringkasan perubahan
	if s.OriginalDBConfigInfo != nil {
		ui.PrintSubHeader("Ringkasan Perubahan : " + s.DBConfigInfo.ConfigName)
		s.printChangeSummary()
		return
	}

	// Default: create flow -> tampilkan ringkasan pembuatan
	ui.PrintSubHeader("Konfigurasi Database Baru: " + s.DBConfigInfo.ConfigName)
	s.printCreateSummary()

}

// printCreateSummary mencetak ringkasan konfigurasi baru.
func (s *Service) printCreateSummary() {
	rows := [][]string{
		{"1", "Nama", s.DBConfigInfo.ConfigName},
		{"2", "Host", s.DBConfigInfo.ServerDBConnection.Host},
		{"3", "Port", fmt.Sprintf("%d", s.DBConfigInfo.ServerDBConnection.Port)},
		{"4", "User", s.DBConfigInfo.ServerDBConnection.User},
	}
	pwState := "(not set)"
	if s.DBConfigInfo.ServerDBConnection.Password != "" {
		pwState = "(set)"
	}
	rows = append(rows, []string{"5", "Password", pwState})

	ui.FormatTable([]string{"No", "Field", "Value"}, rows)
}

// printChangeSummary compares OriginalDBConfigInfo and current DBConfigInfo and prints a short summary.
func (s *Service) printChangeSummary() {
	orig := s.OriginalDBConfigInfo
	if orig == nil {
		ui.PrintInfo("Tidak ada informasi perubahan (tidak ada snapshot asli).")
		return
	}
	rows := [][]string{}
	idx := 1

	// helper to represent password state
	pwState := func(pw string) string {
		if pw == "" {
			return "(not set)"
		}
		return "(set)"
	}

	if orig.ConfigName != s.DBConfigInfo.ConfigName {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), "Nama", orig.ConfigName, s.DBConfigInfo.ConfigName})
		idx++
	}
	if orig.ServerDBConnection.Host != s.DBConfigInfo.ServerDBConnection.Host {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), "Host", orig.ServerDBConnection.Host, s.DBConfigInfo.ServerDBConnection.Host})
		idx++
	}
	if orig.ServerDBConnection.Port != s.DBConfigInfo.ServerDBConnection.Port {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), "Port", fmt.Sprintf("%d", orig.ServerDBConnection.Port), fmt.Sprintf("%d", s.DBConfigInfo.ServerDBConnection.Port)})
		idx++
	}
	if orig.ServerDBConnection.User != s.DBConfigInfo.ServerDBConnection.User {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), "User", orig.ServerDBConnection.User, s.DBConfigInfo.ServerDBConnection.User})
		idx++
	}
	if orig.ServerDBConnection.Password != s.DBConfigInfo.ServerDBConnection.Password {
		rows = append(rows, []string{fmt.Sprintf("%d", idx), "Password", pwState(orig.ServerDBConnection.Password), pwState(s.DBConfigInfo.ServerDBConnection.Password)})
		idx++
	}

	if len(rows) == 0 {
		ui.PrintInfo("Tidak ada perubahan yang terdeteksi pada konfigurasi.")
		return
	}

	// headers: No, Field, Before, After
	ui.FormatTable([]string{"No", "Field", "Before", "After"}, rows)
}

// printShowDetails mencetak seluruh detail konfigurasi dari snapshot (untuk mode show).
func (s *Service) printShowDetails() {
	orig := s.OriginalDBConfigInfo
	if orig == nil {
		ui.PrintInfo("Tidak ada konfigurasi yang dimuat untuk ditampilkan.")
		return
	}

	pwState := "(not set)"
	if orig.ServerDBConnection.Password != "" {
		pwState = "(set)"
	}

	rows := [][]string{
		{"1", "Nama", orig.ConfigName},
		{"2", "File Path", orig.FilePath},
		{"3", "Host", orig.ServerDBConnection.Host},
		{"4", "Port", fmt.Sprintf("%d", orig.ServerDBConnection.Port)},
		{"5", "User", orig.ServerDBConnection.User},
		{"6", "Password", pwState},
		{"8", "File Size", orig.FileSize},
		{"9", "Last Modified", fmt.Sprintf("%v", orig.LastModified)},
		{"10", "Is Valid", fmt.Sprintf("%v", orig.IsValid)},
	}

	ui.FormatTable([]string{"No", "Field", "Value"}, rows)

	// Jika pengguna meminta reveal-password, lakukan konfirmasi dengan meminta
	// ulang encryption password dan coba dekripsi file. Hanya tampilkan password
	// asli jika dekripsi berhasil.
	if s.DBConfigShow != nil && s.DBConfigShow.RevealPassword {
		s.revealPasswordConfirmAndShow(orig)
	}
}

// revealPasswordConfirmAndShow meminta encryption password, mencoba mendekripsi file,
// dan menampilkan password jika dekripsi berhasil.
func (s *Service) revealPasswordConfirmAndShow(orig *structs.DBConfigInfo) {
	if orig.FilePath == "" {
		ui.PrintWarning("Tidak ada file yang terkait untuk memverifikasi password.")
		return
	}

	// Minta ulang encryption key
	key, err := input.AskPassword("Masukkan ulang encryption key untuk verifikasi: ", survey.Required)
	if err != nil {
		ui.PrintWarning("Gagal mendapatkan encryption key: " + err.Error())
		return
	}
	if key == "" {
		ui.PrintWarning("Tidak ada encryption key yang diberikan. Tidak dapat menampilkan password asli.")
		return
	}

	// Gunakan helper untuk memuat dan parse dengan resolver kunci
	info, err := encrypt.LoadAndParseConfig(orig.FilePath, key)
	if err != nil {
		ui.PrintWarning("Enkripsi key salah atau file rusak. Tidak dapat menampilkan password asli.")
		return
	}
	realPw := info.ServerDBConnection.Password

	// Tampilkan result dalam table kecil
	display := "(not set)"
	if realPw != "" {
		display = realPw
	}

	ui.PrintSubHeader("Revealed Password")
	ui.FormatTable([]string{"No", "Field", "Value"}, [][]string{{"1", "Database Password", display}})
}
