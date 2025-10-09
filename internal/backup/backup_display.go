// File : internal/backup/backup_display.go
// Deskripsi : Fungsi utilitas untuk menampilkan informasi backup
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-08
// Last Modified : 2024-10-08

package backup

import (
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/ui"
	"strconv"
)

// DisplayConnectionInfo menampilkan informasi koneksi database sumber
// yang akan di-backup.
func (s *Service) DisplayConnectionInfo(DBConfigInfo structs.DBConfigInfo) {
	ui.PrintSubHeader("Source Database Connection Info")
	// Tampilkan informasi koneksi sumber
	dbconfigInfo := [][]string{
		{"Configuration Name", DBConfigInfo.ConfigName},
		{"Configuration Path", DBConfigInfo.FilePath},
		{"Host", DBConfigInfo.ServerDBConnection.Host},
		{"Port", strconv.Itoa(DBConfigInfo.ServerDBConnection.Port)},
		{"Username", DBConfigInfo.ServerDBConnection.User},
	}
	ui.FormatTable([]string{"Selected Configuration", "Value"}, dbconfigInfo)
}
