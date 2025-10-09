// File : pkg/common/common_helper.go
// Deskripsi : Fungsi utilitas umum yang digunakan di berbagai bagian aplikasi
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package common

import (
	"errors"
	"fmt"
	"path/filepath"
	"sfDBTools/internal/appconfig"
	"strings"

	"github.com/AlecAivazis/survey/v2/terminal"
)

var ErrUserCancelled = errors.New("user_cancelled")

// EnsureConfigExt memastikan nama memiliki suffix .cnf.enc
func EnsureConfigExt(name string) string {
	if strings.HasSuffix(name, ".cnf.enc") {
		return name
	}
	if strings.HasSuffix(name, ".cnf") {
		return name + ".enc"
	}
	return name + ".cnf.enc"
}

// HandleInputError menangani error dari input/survey dan mengubahnya menjadi ErrUserCancelled jika perlu.
func HandleInputError(err error) error {
	if err == terminal.InterruptErr {
		return ErrUserCancelled
	}
	return err
}

// common.TrimConfigSuffix menghapus suffix .cnf.enc dari nama jika ada.
func TrimConfigSuffix(name string) string {
	return strings.TrimSuffix(strings.TrimSuffix(name, ".enc"), ".cnf")
}

// ResolveConfigPath mengubah name/path menjadi absolute path di config dir dan nama normalized tanpa suffix
func ResolveConfigPath(spec string) (string, string, error) {
	if strings.TrimSpace(spec) == "" {
		return "", "", fmt.Errorf("nama atau path file konfigurasi kosong")
	}
	cfg, err := appconfig.LoadConfigFromEnv()
	if err != nil {
		return "", "", fmt.Errorf("gagal memuat konfigurasi aplikasi: %w", err)
	}
	cfgDir := cfg.ConfigDir.DatabaseConfig
	var absPath string
	if filepath.IsAbs(spec) {
		absPath = spec
	} else {
		absPath = filepath.Join(cfgDir, spec)
	}
	absPath = EnsureConfigExt(absPath)

	// Nama normalized
	name := TrimConfigSuffix(filepath.Base(absPath))
	return absPath, name, nil
}
