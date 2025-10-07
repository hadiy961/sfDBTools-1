package appconfig

import (
	"errors"
	"os"

	"github.com/joho/godotenv" // Untuk memuat file .env
	"gopkg.in/yaml.v3"         // Untuk mem-parsing file YAML
)

const configEnvVar = "SFDB_APPS_CONFIG"

// LoadConfigFromEnv memuat file .env (jika ada) dan kemudian
// membaca path konfigurasi dari variabel lingkungan SFDB_APPS_CONFIG.
// Ini mengembalikan struct Config yang terisi, atau error jika gagal.
func LoadConfigFromEnv() (*Config, error) {
	// 1. Memuat variabel lingkungan dari file .env (best practice)
	// Kita abaikan error jika file .env tidak ditemukan, ini wajar
	// di lingkungan produksi di mana variabel env sudah diset.
	_ = godotenv.Load()

	// 2. Membaca path file konfigurasi dari variabel lingkungan
	configPath := os.Getenv(configEnvVar)
	if configPath == "" {
		return nil, errors.New("variabel lingkungan " + configEnvVar + " tidak diset. Mohon tentukan path ke file config.yaml")
	}

	// 3. Memuat dan mem-parsing file konfigurasi
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadConfig membaca dan mem-parsing file YAML dari path yang diberikan
// ke dalam struct Config. Ini adalah fungsi yang reusable.
func LoadConfig(configPath string) (*Config, error) {
	// Membaca konten file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	// Inisialisasi struct Config
	cfg := &Config{}

	// Parsing konten YAML ke dalam struct
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}

	// Best practice: Anda bisa menambahkan logika validasi kustom di sini
	// setelah parsing berhasil (misalnya, cek apakah BaseDirectory tidak kosong)

	return cfg, nil
}
