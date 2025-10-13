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
	ui.Headers("Opsi Backup Database")
	ui.PrintSubHeader("Opsi Backup Database All")
	// menampilkan seluruh opsi backup yang digunakan dalam bentuk table
	headers := []string{"Option", "Value"}
	port := s.BackupAll.BackupOptions.DBConfig.ServerDBConnection.Port
	data := [][]string{
		{"Configuration Path", s.BackupAll.BackupOptions.DBConfig.FilePath},
		{"Host", s.BackupAll.BackupOptions.DBConfig.ServerDBConnection.Host},
		{"Port", strconv.Itoa(port)},
		{"Username", s.BackupAll.BackupOptions.DBConfig.ServerDBConnection.User},
		{"Backup Directory", s.BackupAll.BackupOptions.OutputDirectory},
		{"Compression", s.BackupAll.BackupOptions.Compression.Type},
		{"Compression Level", (s.BackupAll.BackupOptions.Compression.Level)},
		{"Encryption Enabled", strconv.FormatBool(s.BackupAll.BackupOptions.Encryption.Enabled)},
		{"Cleanup Enabled", strconv.FormatBool(s.BackupAll.Cleanup.Enabled)},
		{"Cleanup Schedule", s.BackupAll.Cleanup.Scheduled},
		{"Retention Days", strconv.Itoa(s.BackupAll.Cleanup.RetentionDays)},
		{"Exclude Databases", ui.FormatStringSlice(s.BackupAll.Exclude.Databases)},
		{"Exclude Users", strconv.FormatBool(s.BackupAll.Exclude.Users)},
		{"Exclude System Databases", strconv.FormatBool(s.BackupAll.Exclude.SystemsDB)},
		{"Exclude Data", strconv.FormatBool(s.BackupAll.Exclude.Data)},
		{"Database List File", s.BackupAll.DBList.File},
		{"Verification Disk Check", strconv.FormatBool(s.BackupAll.BackupOptions.DiskCheck)},
		{"Capture GTID", strconv.FormatBool(s.BackupAll.CaptureGtid)},
	}
	ui.FormatTable(headers, data)
}
