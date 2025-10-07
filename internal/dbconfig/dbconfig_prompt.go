// File : internal/dbconfig/dbconfig_prompt.go
// Deskripsi : Logika untuk prompt input user pada modul dbconfig
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package dbconfig

import (
	"os"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/common"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"

	"github.com/AlecAivazis/survey/v2"
)

func (s *Service) promptDBConfigInfo() error {
	ui.PrintSubHeader("Please provide the following database configuration details:")
	var err error

	s.DBConfigInfo.ServerDBConnection.Host, err = input.AskString("Database Host", s.DBConfigInfo.ServerDBConnection.Host, survey.Required)
	if err != nil {
		return common.HandleInputError(err)
	}

	s.DBConfigInfo.ServerDBConnection.Port, err = input.AskInt("Database Port", s.DBConfigInfo.ServerDBConnection.Port, survey.Required)
	if err != nil {
		return common.HandleInputError(err)
	}

	s.DBConfigInfo.ServerDBConnection.User, err = input.AskString("Database User", s.DBConfigInfo.ServerDBConnection.User, survey.Required)
	if err != nil {
		return common.HandleInputError(err)
	}

	//Cek password dari env SFDB_DB_PASSWORD
	//Jika tidak ada, beri tahu user untuk mengaturnya di env
	//Untuk keamanan, jangan minta input password di prompt
	//Namun, jika ingin minta input, uncomment kode di bawah dan comment kode pengecekan env
	// s.Logger.Debug("Mengecek environment variable SFDB_DB_PASSWORD untuk password database...")
	s.DBConfigInfo.ServerDBConnection.Password = os.Getenv(common.SFDB_DB_PASSWORD)
	if s.DBConfigInfo.ServerDBConnection.Password == "" {
		ui.PrintWarning("Environment variable SFDB_DB_PASSWORD tidak ditemukan atau kosong. Silakan atur SFDB_DB_PASSWORD atau ketik password.")
	}
	// Allow empty password in edit flow to mean "keep existing password".
	// Detect edit flow if we already have an OriginalDBConfigInfo or OriginalConfigName.
	isEditFlow := s.OriginalDBConfigInfo != nil || s.OriginalConfigName != ""

	// First try accepting empty input (validator nil). If this is create flow and
	// user entered empty, we enforce non-empty by asking again with Required.
	pw, err := input.AskPassword("Database Password", nil)
	if err != nil {
		return common.HandleInputError(err)
	}

	if pw == "" {
		if isEditFlow {
			// keep existing password (from env or previously loaded file)
			// do nothing; leave s.DBConfigInfo.ServerDBConnection.Password as-is
		} else {
			// create flow: password required
			pw, err = input.AskPassword("Database Password", survey.Required)
			if err != nil {
				return common.HandleInputError(err)
			}
			s.DBConfigInfo.ServerDBConnection.Password = pw
		}
	} else {
		// user provided a new password -> overwrite
		s.DBConfigInfo.ServerDBConnection.Password = pw
	}

	return nil
}

// PERBAIKAN: Fungsi ini juga sekarang hanya mengembalikan error.
func (s *Service) promptDBConfigName(mode string) error {
	if s.DBConfigInfo.ConfigName == "" {
		s.DBConfigInfo.ConfigName = "my_database_config" // Set default jika kosong
	}
	ui.PrintSubHeader("Please provide the configuration name:")

	// Mulai loop untuk meminta input sampai valid
	for {
		// 1. Minta input dari pengguna
		nameValidator := input.ComposeValidators(survey.Required, input.ValidateFilename)
		configName, err := input.AskString("Configuration Name", s.DBConfigInfo.ConfigName, nameValidator)
		if err != nil {
			return common.HandleInputError(err) // Keluar jika pengguna membatalkan (misal: Ctrl+C)
		}

		// Selalu update nama di struct dengan versi dinormalisasi (trim suffix)
		s.DBConfigInfo.ConfigName = common.TrimConfigSuffix(configName)

		// 2. Lakukan validasi keunikan nama
		err = s.CheckConfigurationNameUnique(mode)
		if err != nil {
			// 3. JIKA TIDAK UNIK: Tampilkan error dan loop akan berulang
			ui.PrintError(err.Error())
			continue // Lanjutkan ke iterasi loop berikutnya
		}

		// 4. JIKA UNIK: Keluar dari loop
		break
	}

	// Tampilkan informasi akhir
	ui.PrintInfo("Konfigurasi akan disimpan sebagai : " + buildFileName(s.DBConfigInfo.ConfigName))
	return nil
}

// promptSelectExistingConfig menampilkan daftar file konfigurasi dari direktori
// konfigurasi aplikasi dan meminta pengguna memilih salah satu.
func (s *Service) promptSelectExistingConfig() error {
	info, err := encrypt.SelectExistingDBConfig("Select Existing Configuration File")
	if err != nil {
		return err
	}

	// Muat data ke struct DBConfigInfo
	s.DBConfigInfo = &structs.DBConfigInfo{
		FilePath:           info.FilePath,
		ConfigName:         info.ConfigName,
		ServerDBConnection: info.ServerDBConnection,
		FileSize:           info.FileSize,
		LastModified:       info.LastModified,
	}

	// Setelah berhasil memuat isi file, simpan snapshot data asli agar dapat
	// dibandingkan dengan perubahan yang dilakukan user. Sertakan metadata file.
	s.OriginalDBConfigInfo = &structs.DBConfigInfo{
		FilePath:           info.FilePath,
		ConfigName:         info.ConfigName,
		ServerDBConnection: info.ServerDBConnection,
		FileSize:           info.FileSize,
		LastModified:       info.LastModified,
	}

	ui.PrintInfo("Memuat konfigurasi dari: " + info.FilePath + " Name: " + info.ConfigName + " (Last Modified: " + info.LastModified.String() + ", Size: " + info.FileSize + ")")
	return nil
}
