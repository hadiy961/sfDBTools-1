// File : internal/backup/backup_main.go
// Deskripsi : Service utama untuk modul backup
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-08
// Last Modified : 2024-10-08

package backup

import (
	config "sfDBTools/internal/appconfig"
	log "sfDBTools/internal/applog"
	"sfDBTools/internal/structs"
)

// Service adalah layanan inti yang menjalankan logika dbconfig.
type Service struct {
	Logger        log.Logger
	Config        *config.Config
	BackupDB      *structs.BackupDBFlags
	BackupAll     *structs.BackupAllFlags
	BackupInfo    *structs.BackupInfo
	BackupOptions *structs.BackupOptions
	DBConfigInfo  *structs.DBConfigInfo
	FilterInfo    *structs.FilterInfo  // Informasi statistik filtering database
	FilterStats   *DatabaseFilterStats // Statistik filtering database
}

// NewService membuat instance baru dari Service dengan dependensi yang di-inject.
func NewService(logger log.Logger, cfg *config.Config, dbConfig interface{}) *Service {
	svc := &Service{
		Logger: logger,
		Config: cfg,
	}

	// Jika caller memberikan flags (bisa berupa BackupAllFlags atau BackupDBFlags),
	// ambil nilai DBConfigInfo dan flag Interactive.
	if dbConfig != nil {
		switch v := dbConfig.(type) {
		case *structs.BackupAllFlags:
			svc.BackupAll = v
			svc.BackupInfo = &v.BackupInfo
			svc.BackupOptions = &v.BackupOptions
			svc.DBConfigInfo = &v.BackupOptions.DBConfig
			svc.DBConfigInfo.ServerDBConnection = v.BackupOptions.DBConfig.ServerDBConnection
		case *structs.BackupDBFlags:
			svc.BackupDB = v
			svc.BackupInfo = &v.BackupInfo
			svc.BackupOptions = &v.BackupOptions
			svc.DBConfigInfo = &v.BackupOptions.DBConfig
			svc.DBConfigInfo.ServerDBConnection = v.BackupOptions.DBConfig.ServerDBConnection
		default:
			// Unknown type: buat default kosong
			svc.BackupInfo = &structs.BackupInfo{}
			svc.BackupOptions = &structs.BackupOptions{}
			svc.DBConfigInfo = &structs.DBConfigInfo{}
		}
	} else {
		svc.BackupInfo = &structs.BackupInfo{}
		svc.BackupOptions = &structs.BackupOptions{}
		svc.DBConfigInfo = &structs.DBConfigInfo{}
	}

	return svc
}
