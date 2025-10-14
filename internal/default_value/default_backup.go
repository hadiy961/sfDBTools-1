// File : internal/default_value/default_backup.go
// Deskripsi : Nilai default untuk flags pada modul backup
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-08
// Last Modified : 2024-10-08

package defaultvalue

import (
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/structs"
)

// GetDefaultBackupFlags returns default values for BackupFlags
func GetDefaultBackupFlags() (*structs.BackupDBFlags, error) {
	// Muat konfigurasi aplikasi untuk mendapatkan direktori konfigurasi
	cfg, err := appconfig.LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}
	return &structs.BackupDBFlags{
		BackupOptions: structs.BackupOptions{
			Encryption: structs.EncryptionOptions{
				Enabled: cfg.Backup.Encryption.Enabled,
				Key:     cfg.Backup.Encryption.Key,
			},
			Compression: structs.CompressionOptions{
				Type:    cfg.Backup.Compression.Type,
				Level:   cfg.Backup.Compression.Level,
				Enabled: cfg.Backup.Compression.Required,
			},
			Cleanup: structs.CleanupOptions{
				Enabled:       cfg.Backup.Retention.CleanupEnabled,
				Scheduled:     cfg.Backup.Retention.CleanupSchedule,
				RetentionDays: cfg.Backup.Retention.Days,
			},
			Exclude: structs.ExcludeOptions{
				Databases: cfg.Backup.Exclude.Databases,
				Users:     cfg.Backup.Exclude.User,
				SystemsDB: cfg.Backup.Exclude.SystemDatabases,
				Data:      cfg.Backup.Exclude.Data,
			},
			OutputDirectory: cfg.Backup.Output.BaseDirectory,
			DiskCheck:       cfg.Backup.Verification.DiskSpaceCheck,
			DBList:          cfg.Backup.DBList.File,
		},
		BackupInfo: structs.BackupInfo{
			Enabled: cfg.Backup.Output.CreateBackupInfo,
		},
	}, nil
}

// GetDefaultBackupAllFlags returns default values for BackupAllFlags
func GetDefaultBackupAllFlags() (*structs.BackupAllFlags, error) {
	// Muat konfigurasi aplikasi untuk mendapatkan direktori konfigurasi
	cfg, err := appconfig.LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}
	return &structs.BackupAllFlags{
		BackupOptions: structs.BackupOptions{
			Encryption: structs.EncryptionOptions{
				Enabled: cfg.Backup.Encryption.Enabled,
				Key:     cfg.Backup.Encryption.Key,
			},
			Compression: structs.CompressionOptions{
				Type:    cfg.Backup.Compression.Type,
				Level:   cfg.Backup.Compression.Level,
				Enabled: cfg.Backup.Compression.Required,
			},
			Cleanup: structs.CleanupOptions{
				Enabled:       cfg.Backup.Retention.CleanupEnabled,
				Scheduled:     cfg.Backup.Retention.CleanupSchedule,
				RetentionDays: cfg.Backup.Retention.Days,
			},

			Exclude: structs.ExcludeOptions{
				Databases: cfg.Backup.Exclude.Databases,
				Users:     cfg.Backup.Exclude.User,
				SystemsDB: cfg.Backup.Exclude.SystemDatabases,
				Data:      cfg.Backup.Exclude.Data,
			},
			OutputDirectory: cfg.Backup.Output.BaseDirectory,
			DiskCheck:       cfg.Backup.Verification.DiskSpaceCheck,
			DBList:          cfg.Backup.DBList.File,
		},
		BackupInfo: structs.BackupInfo{
			Enabled: cfg.Backup.Output.CreateBackupInfo,
		},
		CaptureGtid: cfg.Backup.Output.CaptureGtid,
	}, nil
}
