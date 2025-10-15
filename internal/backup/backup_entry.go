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

// BackupConfig untuk konfigurasi backup entry point
type BackupEntryConfig struct {
	HeaderTitle string
	ShowOptions bool
	BackupMode  string // "separate" atau "combined"
	EnableGTID  bool   // apakah perlu capture GTID
	SuccessMsg  string
	LogPrefix   string
}

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

	// Log database yang akan di-backup (jika mode separate)
	if config.BackupMode == "separate" {
		s.Logger.Info("Database yang akan di-backup:")
		for _, db := range dbFiltered {
			s.Logger.Infof("- %s", db)
		}
	}

	// Lakukan backup dengan mode yang ditentukan
	if err := s.ExecuteBackupWithDetailCollection(ctx, client, dbFiltered, config.BackupMode); err != nil {
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
		HeaderTitle: "Backup Database",
		ShowOptions: false,
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
		HeaderTitle: "Backup Semua Database",
		ShowOptions: true,
		BackupMode:  "combined",
		EnableGTID:  true,
		SuccessMsg:  "Proses backup semua database selesai.",
		LogPrefix:   "Proses backup semua database",
	}
	return s.ExecuteBackupCommand(config)
}
