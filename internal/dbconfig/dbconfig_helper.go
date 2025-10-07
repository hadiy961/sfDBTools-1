// File : internal/dbconfig/dbconfig_helper.go
// Deskripsi : Helper functions untuk modul dbconfig
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03

package dbconfig

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/common"
	"sfDBTools/pkg/encrypt"
	"strings"
)

// ErrUserCancelled adalah sentinel error untuk menandai pembatalan oleh pengguna.
var ErrUserCancelled = errors.New("user_cancelled")

// buildFileName menormalkan input (menghapus suffix jika ada) lalu memastikan suffix .cnf.enc
// Tujuan: menghindari duplikasi suffix saat user sudah mengetikkan nama dengan ekstensi.
func buildFileName(name string) string {
	return common.EnsureConfigExt(common.TrimConfigSuffix(strings.TrimSpace(name)))
}

// filePathInConfigDir membangun absolute path di dalam config dir untuk nama file konfigurasi yang diberikan.
func (s *Service) filePathInConfigDir(name string) string {
	cfgDir := s.Config.ConfigDir.DatabaseConfig
	return filepath.Join(cfgDir, buildFileName(name))
}

// resolveConfigPath mengubah name/path menjadi absolute path di config dir dan nama normalized tanpa suffix
func (s *Service) resolveConfigPath(spec string) (string, string, error) {
	if strings.TrimSpace(spec) == "" {
		return "", "", fmt.Errorf("nama atau path file konfigurasi kosong")
	}
	cfgDir := s.Config.ConfigDir.DatabaseConfig
	var absPath string
	if filepath.IsAbs(spec) {
		absPath = spec
	} else {
		absPath = filepath.Join(cfgDir, spec)
	}
	absPath = common.EnsureConfigExt(absPath)

	// Nama normalized
	name := common.TrimConfigSuffix(filepath.Base(absPath))
	return absPath, name, nil
}

// loadSnapshotFromPath membaca file terenkripsi, mencoba dekripsi (jika kunci tersedia/di-prompt),
// parse nilai penting, dan mengisi s.OriginalDBConfigInfo beserta metadata.
func (s *Service) loadSnapshotFromPath(absPath string) error {
	info, err := encrypt.LoadAndParseConfig(absPath, s.DBConfigInfo.EncryptionKey)
	if err != nil {
		s.fillOriginalInfoFromMeta(absPath, structs.DBConfigInfo{})
		return err
	}
	s.fillOriginalInfoFromMeta(absPath, *info)
	return nil
}

// fillOriginalInfoFromMeta mengisi OriginalDBConfigInfo dengan metadata file dan nilai koneksi yang tersedia
func (s *Service) fillOriginalInfoFromMeta(absPath string, info structs.DBConfigInfo) {
	var fileSizeStr string
	var lastMod = info.LastModified
	if fi, err := os.Stat(absPath); err == nil {
		fileSizeStr = fmt.Sprintf("%d bytes", fi.Size())
		lastMod = fi.ModTime()
	}

	s.OriginalDBConfigInfo = &structs.DBConfigInfo{
		FilePath:           absPath,
		ConfigName:         common.TrimConfigSuffix(filepath.Base(absPath)),
		ServerDBConnection: info.ServerDBConnection,
		FileSize:           fileSizeStr,
		LastModified:       lastMod,
	}
}
