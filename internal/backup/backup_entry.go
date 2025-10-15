// File : internal/backup/backup_entry.go
// Deskripsi : Entry points untuk semua jenis backup database
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-15
// Last Modified : 2024-10-15

package backup

import (
	"context"
	"sfDBTools/pkg/ui"
)

// ExecuteBackupCommand adalah unified entry point untuk semua jenis backup
func (s *Service) ExecuteBackupCommand(config BackupEntryConfig) error {
	ctx := context.Background()

	// Setup session (koneksi database, filter database, dll)
	client, dbFiltered, originalMaxStatementsTime, err := s.PrepareBackupSession(ctx, config.HeaderTitle, config.ShowOptions)
	if err != nil {
		return err
	}
	//Pastikan koneksi ditutup di akhir
	defer client.Close()

	// Check flag capture GTID jika diaktifkan
	if config.EnableGTID {
		if err := s.CaptureGTIDIfNeeded(ctx, client); err != nil {
			s.Logger.Warn("Gagal menangani opsi capture GTID: " + err.Error())
			s.KembalikanMaxStatementsTime(ctx, client, originalMaxStatementsTime)
			return err
		}
	}

	// Lakukan backup dengan mode yang ditentukan
	if err := s.ExecuteBackup(ctx, client, dbFiltered, config.BackupMode, true); err != nil {
		s.Logger.Error(config.LogPrefix + " gagal: " + err.Error())
		// Kembalikan nilai awal max_statement_time jika ada error
		s.KembalikanMaxStatementsTime(ctx, client, originalMaxStatementsTime)
		return err
	}

	// Kembalikan nilai awal max_statement_time
	s.KembalikanMaxStatementsTime(ctx, client, originalMaxStatementsTime)

	// Print success message jika ada
	if config.SuccessMsg != "" {
		ui.PrintSuccess(config.SuccessMsg)
	}

	return nil
}

// BackupDatabase melakukan backup database dengan file terpisah per database
func (s *Service) BackupDatabase() error {
	config := BackupEntryConfig{
		HeaderTitle: "Backup Database (Tiap Database Terpisah)",
		ShowOptions: true,
		BackupMode:  "separate",
		EnableGTID:  false,
		SuccessMsg:  "", // No success message for database backup
		LogPrefix:   "Proses backup database",
	}
	return s.ExecuteBackupCommand(config)
}

// BackupAllDatabases melakukan backup semua database dalam satu file
func (s *Service) BackupAllDatabases() error {
	config := BackupEntryConfig{
		HeaderTitle: "Backup Database (Tiap Database Digabung)",
		ShowOptions: true,
		BackupMode:  "combined",
		EnableGTID:  true,
		SuccessMsg:  "Proses backup semua database selesai.",
		LogPrefix:   "Proses backup semua database",
	}
	return s.ExecuteBackupCommand(config)
}
