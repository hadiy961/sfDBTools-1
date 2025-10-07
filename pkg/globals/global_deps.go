// File : pkg/globals/global_deps.go
// Deskripsi : Definisi struct untuk menyimpan dependensi global aplikasi
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03

package globals

// Kita harus mendefinisikan tipe data di sini untuk menghindari import cycle
// Gunakan path config dan log yang benar di sini
import (
	config "sfDBTools/internal/appconfig"
	log "sfDBTools/internal/applog"
)

// Dependencies adalah struct yang menyimpan semua dependensi global aplikasi.
type Dependencies struct {
	Config *config.Config
	Logger log.Logger
}

// Global variable untuk menyimpan dependensi yang di-inject
var Deps *Dependencies

// GetLogger adalah helper untuk mengakses logger dari package lain
func GetLogger() log.Logger {
	if Deps == nil {
		return nil
	}
	return Deps.Logger
}

// GetConfig adalah helper untuk mengakses config dari package lain
func GetConfig() *config.Config {
	if Deps == nil {
		return nil
	}
	return Deps.Config
}
