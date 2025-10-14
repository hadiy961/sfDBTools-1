// File : internal/encrypt/encrypt_main.go
// Deskripsi : Service utama untuk modul encrypt/decrypt
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-14
// Last Modified : 2024-10-14

package encrypt

import (
	config "sfDBTools/internal/appconfig"
	log "sfDBTools/internal/applog"
	"sfDBTools/internal/structs"
)

// Service adalah layanan inti yang menjalankan logika encrypt/decrypt.
type Service struct {
	Logger      log.Logger
	Config      *config.Config
	EncryptInfo *structs.EncryptFlags
	DecryptInfo *structs.DecryptFlags
}

// NewService membuat instance baru dari Service dengan dependensi yang di-inject.
func NewService(logger log.Logger, cfg *config.Config, flags interface{}) *Service {
	svc := &Service{
		Logger: logger,
		Config: cfg,
	}

	// Set flags berdasarkan tipe yang diberikan
	if flags != nil {
		switch v := flags.(type) {
		case *structs.EncryptFlags:
			svc.EncryptInfo = v
		case *structs.DecryptFlags:
			svc.DecryptInfo = v
		}
	}

	return svc
}
