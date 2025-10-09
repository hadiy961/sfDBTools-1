package backup

import (
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/ui"
)

// promptSelectExistingConfig menampilkan daftar file konfigurasi dari direktori
// konfigurasi aplikasi dan meminta pengguna memilih salah satu.
func (s *Service) promptSelectExistingConfig() error {
	info, err := encrypt.SelectExistingDBConfig("Select Existing Configuration File")
	if err != nil {
		return err
	}

	// Muat data ke struct DBConfigInfo
	// s.BackupAll. = &structs.DBConfigInfo{
	// 	FilePath:           info.FilePath,
	// 	ConfigName:         info.ConfigName,
	// 	ServerDBConnection: info.ServerDBConnection,
	// 	FileSize:           info.FileSize,
	// 	LastModified:       info.LastModified,
	// }

	ui.PrintInfo("Memuat konfigurasi dari: " + info.FilePath + " Name: " + info.ConfigName + " (Last Modified: " + info.LastModified.String() + ", Size: " + info.FileSize + ")")
	return nil
}
