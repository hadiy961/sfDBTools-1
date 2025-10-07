// File : internal/dbconfig/dbconfig_main.go
// Deskripsi : Service utama untuk modul dbconfig
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package dbconfig

import (
	config "sfDBTools/internal/appconfig"
	log "sfDBTools/internal/applog"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/common"
)

// Service adalah layanan inti yang menjalankan logika dbconfig.
type Service struct {
	Logger         log.Logger
	Config         *config.Config
	DBConfigCreate *structs.DBConfigCreateFlags
	DBConfigEdit   *structs.DBConfigEditFlags
	DBConfigInfo   *structs.DBConfigInfo
	DBConfigShow   *structs.DBConfigShowFlags
	// OriginalConfigName menyimpan nama file konfigurasi yang dibuka untuk mode edit.
	OriginalConfigName string
	// OriginalDBConfigInfo menyimpan salinan data konfigurasi sebelum diedit (jika tersedia)
	OriginalDBConfigInfo *structs.DBConfigInfo
}

// NewService membuat instance baru dari Service dengan dependensi yang di-inject.
func NewService(logger log.Logger, cfg *config.Config, dbConfig interface{}) *Service {
	svc := &Service{
		Logger: logger,
		Config: cfg,
	}

	// Jika caller memberikan flags (bisa berupa DBConfigCreateFlags atau DBConfigEditFlags),
	// ambil nilai DBConfigInfo dan flag Interactive.
	if dbConfig != nil {
		switch v := dbConfig.(type) {
		case *structs.DBConfigCreateFlags:
			svc.DBConfigCreate = v
			svc.DBConfigInfo = &v.DBConfigInfo
		case *structs.DBConfigEditFlags:
			svc.DBConfigEdit = v
			svc.DBConfigInfo = &structs.DBConfigInfo{
				ConfigName:         v.File,
				ServerDBConnection: v.DBConfigInfo.ServerDBConnection,
				EncryptionKey:      v.DBConfigInfo.EncryptionKey,
			}
			// Simpan nama asli file yang diberikan dalam flag sebagai original (normalized)
			svc.OriginalConfigName = common.TrimConfigSuffix(v.File)
		case *structs.DBConfigShowFlags:
			svc.DBConfigShow = v
			svc.DBConfigInfo = &structs.DBConfigInfo{
				FilePath:      v.File,
				EncryptionKey: v.EncryptionKey, // Akan diisi dari env atau flag lain jika diperlukan
			}
		default:
			// Unknown type: buat default kosong
			svc.DBConfigInfo = &structs.DBConfigInfo{}
		}
	} else {
		svc.DBConfigInfo = &structs.DBConfigInfo{}
	}

	return svc
}
