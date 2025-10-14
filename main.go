package main

import (
	"fmt"
	"os"
	"sfDBTools/cmd"
	config "sfDBTools/internal/appconfig"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/globals"

	applog "sfDBTools/internal/applog"
)

// Inisialisasi awal untuk Config dan Logger.
var cfg *config.Config
var appLogger applog.Logger

func main() {
	// 1. Muat Konfigurasi
	var err error
	cfg, err = config.LoadConfigFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: Gagal memuat konfigurasi: %v\n", err)
		os.Exit(1)
	}

	// 2. Inisialisasi Logger Kustom
	appLogger = applog.NewLogger()

	// 3. Inisialisasi koneksi database dari environment variables
	dbClient, err := database.InitializeDatabaseFromEnv()
	if err != nil {
		// Log error tapi jangan exit, karena tidak semua command membutuhkan database
		appLogger.Warn(fmt.Sprintf("Gagal menginisialisasi koneksi database: %v", err))
		appLogger.Info("Aplikasi akan berjalan tanpa koneksi database aktif")
	} else {
		appLogger.Info("Koneksi database berhasil diinisialisasi")
		// Tutup koneksi database saat aplikasi selesai
		defer func() {
			if closeErr := dbClient.Close(); closeErr != nil {
				appLogger.Warn(fmt.Sprintf("Gagal menutup koneksi database: %v", closeErr))
			} else {
				appLogger.Info("Koneksi database berhasil ditutup")
			}
		}()
	}

	// 4. Buat objek dependensi untuk di-inject
	deps := &globals.Dependencies{
		Config:   cfg,
		Logger:   appLogger,
		DBClient: dbClient, // Tambahkan database client ke dependencies
	}

	// 5. Jalankan perintah Cobra dengan dependensi
	cmd.Execute(deps)
}
