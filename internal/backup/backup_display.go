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

// DisplayBackupOptions menampilkan opsi backup yang digunakan
func (s *Service) DisplayBackupAllOptions() {
	ui.PrintSubHeader("Opsi Backup Database All")
	// menampilkan seluruh opsi backup yang digunakan dalam bentuk table
	headers := []string{"Option", "Value"}
	port := s.DBConfigInfo.ServerDBConnection.Port
	data := [][]string{
		{"Configuration Path", s.DBConfigInfo.FilePath},
		{"Host", s.DBConfigInfo.ServerDBConnection.Host},
		{"Port", strconv.Itoa(port)},
		{"Username", s.DBConfigInfo.ServerDBConnection.User},
		{"Backup Directory", s.BackupOptions.OutputDirectory},
		{"Compression", s.BackupOptions.Compression.Type},
		{"Compression Level", (s.BackupOptions.Compression.Level)},
		{"Encryption Enabled", strconv.FormatBool(s.BackupOptions.Encryption.Enabled)},
		{"Cleanup Enabled", strconv.FormatBool(s.BackupOptions.Cleanup.Enabled)},
		{"Cleanup Schedule", s.BackupOptions.Cleanup.Scheduled},
		{"Retention Days", strconv.Itoa(s.BackupOptions.Cleanup.RetentionDays)},
		{"Exclude Databases", ui.FormatStringSlice(s.BackupOptions.Exclude.Databases)},
		{"Exclude Users", strconv.FormatBool(s.BackupOptions.Exclude.Users)},
		{"Exclude System Databases", strconv.FormatBool(s.BackupOptions.Exclude.SystemsDB)},
		{"Exclude Data", strconv.FormatBool(s.BackupOptions.Exclude.Data)},
		{"Database List File", s.BackupOptions.DBList},
		{"Verification Disk Check", strconv.FormatBool(s.BackupOptions.DiskCheck)},
		{"Capture GTID", strconv.FormatBool(s.BackupAll.CaptureGtid)},
	}
	ui.FormatTable(headers, data)
}
