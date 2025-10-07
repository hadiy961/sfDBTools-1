package main

import (
	"fmt"
	"os"
	"sfDBTools/cmd"
	config "sfDBTools/internal/appconfig"
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

	// 3. Buat objek dependensi untuk di-inject
	deps := &globals.Dependencies{
		Config: cfg,
		Logger: appLogger,
	}

	// 4. Jalankan perintah Cobra dengan dependensi
	cmd.Execute(deps)
}
