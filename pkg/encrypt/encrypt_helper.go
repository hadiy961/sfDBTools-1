package encrypt

import (
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/common"
	"sfDBTools/pkg/fs"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/parsing"
	"sfDBTools/pkg/ui"
	"strings"
)

// LoadAndParseConfig membaca file terenkripsi, mendapatkan kunci (jika tidak diberikan),
// mendekripsi, dan mem-parsing isi INI menjadi DBConfigInfo (tanpa metadata file).
func LoadAndParseConfig(absPath string, key string) (*structs.DBConfigInfo, error) {
	// Ambil kunci enkripsi dari argumen, jika kosong minta dari user
	// Gunakan nilai default dari structs.DBConfigInfo jika ada
	EncryptionKey := structs.DBConfigInfo{}.EncryptionKey
	// Baca file
	data, err := fs.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca file konfigurasi: %w", err)
	}

	// Dapatkan kunci enkripsi (jika tidak diberikan)
	// Gunakan helper yang sudah ada untuk konsistensi
	k := strings.TrimSpace(key)
	if k == "" {
		var src string
		// Jika kunci tidak diberikan, minta dari env atau prompt
		k, src, err = ResolveEncryptionKey(EncryptionKey)
		_ = src // source tidak digunakan di sini
		if err != nil {
			return nil, fmt.Errorf("kunci enkripsi tidak tersedia: %w", err)
		}
	}

	// Dekripsi konten
	plaintext, err := DecryptAES(data, []byte(k))
	if err != nil {
		return nil, fmt.Errorf("gagal mendekripsi file: %w", err)
	}

	// Parsing INI (gunakan helper parsing yang sudah ada)
	// Hasil parsing adalah map[string]string, kita perlu mapping ke DBConfigInfo
	// Parsing hanya bagian [client]
	parsed := parsing.ParseINIClient(string(plaintext))
	info := &structs.DBConfigInfo{ConfigName: common.TrimConfigSuffix(filepath.Base(absPath))}
	if parsed != nil {
		if h, ok := parsed["host"]; ok {
			info.ServerDBConnection.Host = h
		}
		if p, ok := parsed["port"]; ok {
			fmt.Sscanf(p, "%d", &info.ServerDBConnection.Port)
		}
		if u, ok := parsed["user"]; ok {
			info.ServerDBConnection.User = u
		}
		if pw, ok := parsed["password"]; ok {
			info.ServerDBConnection.Password = pw
		}
	}
	return info, nil
}

func SelectExistingDBConfig(purpose string) (structs.DBConfigInfo, error) {
	// Tujuan: menampilkan daftar file konfigurasi yang ada dan membiarkan user memilih satu
	// Kembalikan DBConfigInfo yang berisi metadata file dan detail koneksi (jika berhasil dimuat)
	// Jika tidak ada file, kembalikan error
	// Jika user membatalkan, kembalikan error khusus
	ui.PrintSubHeader(purpose)

	// Muat konfigurasi aplikasi untuk mendapatkan direktori konfigurasi
	cfg, err := appconfig.LoadConfigFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: Gagal memuat konfigurasi: %v\n", err)
		os.Exit(1)
	}

	// Baca isi direktori konfigurasi
	configDir := cfg.ConfigDir.DatabaseConfig
	DBConfig := structs.DBConfigInfo{}

	// Bacaan direktori
	files, err := fs.ReadDirFiles(configDir)
	if err != nil {
		return DBConfig, fmt.Errorf("gagal membaca direktori konfigurasi '%s': %w", configDir, err)
	}

	// filter hanya *.cnf.enc (gunakan helper EnsureConfigExt untuk konsistensi)
	filtered := make([]string, 0, len(files))
	for _, f := range files {
		if common.EnsureConfigExt(f) == f { // artinya sudah ber-ekstensi .cnf.enc
			filtered = append(filtered, f)
		}
	}

	// Jika tidak ada file, kembalikan error
	if len(filtered) == 0 {
		ui.PrintWarning("Tidak ditemukan file konfigurasi di direktori: " + configDir)
		ui.PrintInfo("Silakan buat file konfigurasi baru terlebih dahulu dengan perintah 'dbconfig create'.")
		return DBConfig, fmt.Errorf("tidak ada file konfigurasi untuk dipilih")
	}

	// Buat opsi dengan nama file
	options := make([]string, 0, len(filtered))
	options = append(options, filtered...)

	// Tampilkan menu dan dapatkan pilihan
	idx, err := input.ShowMenu("Pilih file konfigurasi :", options)
	if err != nil {
		return DBConfig, common.HandleInputError(err)
	}

	// index adalah 1-based
	selected := options[idx-1]
	name := common.TrimConfigSuffix(selected)

	// Coba baca file yang dipilih
	filePath := filepath.Join(configDir, selected)
	// Coba muat file yang dipilih (read+decrypt+parse ditangani helper)
	// Gunakan helper untuk memuat dan mem-parsing isi
	info, err := LoadAndParseConfig(filePath, DBConfig.EncryptionKey)
	if err != nil {
		return DBConfig, err
	} else if info != nil {
		DBConfig.ServerDBConnection = info.ServerDBConnection
	}

	// Ambil metadata file (size dan last modified)
	var fileSizeStr string
	var lastModTime = DBConfig.LastModified
	if fi, err := os.Stat(filePath); err == nil {
		fileSizeStr = fmt.Sprintf("%d bytes", fi.Size())
		lastModTime = fi.ModTime()
	}

	// Setelah berhasil memuat isi file, simpan snapshot data asli agar dapat
	// dibandingkan dengan perubahan yang dilakukan user. Sertakan metadata file.
	DBConfig.FilePath = filePath
	DBConfig.ConfigName = name
	DBConfig.ServerDBConnection = info.ServerDBConnection
	DBConfig.FileSize = fileSizeStr
	DBConfig.LastModified = lastModTime

	return DBConfig, nil
}
