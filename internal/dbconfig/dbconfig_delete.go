// File : internal/dbconfig/dbconfig_delete.go
// Deskripsi : Logika untuk menghapus file konfigurasi database
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03

package dbconfig

import (
	"fmt"
	"path/filepath"
	"sfDBTools/pkg/common"
	"sfDBTools/pkg/fs"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
)

// PromptDeleteConfigs shows multi-select of existing config files and deletes selected ones after confirmation.
func (s *Service) PromptDeleteConfigs() error {
	ui.Headers("Delete Database Configurations")

	configDir := s.Config.ConfigDir.DatabaseConfig

	files, err := fs.ReadDirFiles(configDir)
	if err != nil {
		return fmt.Errorf("gagal membaca direktori konfigurasi: %w", err)
	}
	// filter hanya *.cnf.enc (gunakan helper ensureConfigExt untuk konsistensi)
	filtered := make([]string, 0, len(files))
	for _, f := range files {
		if common.EnsureConfigExt(f) == f { // artinya sudah ber-ekstensi .cnf.enc
			filtered = append(filtered, f)
		}
	}
	if len(filtered) == 0 {
		ui.PrintInfo("Tidak ada file konfigurasi untuk dihapus.")
		return nil
	}

	idxs, err := input.ShowMultiSelect("Pilih file konfigurasi yang akan dihapus:", filtered)
	if err != nil {
		return common.HandleInputError(err)
	}

	selected := make([]string, 0, len(idxs))
	for _, i := range idxs {
		if i >= 1 && i <= len(filtered) {
			selected = append(selected, filepath.Join(configDir, filtered[i-1]))
		}
	}

	if len(selected) == 0 {
		ui.PrintInfo("Tidak ada file terpilih untuk dihapus.")
		return nil
	}

	ok, err := input.AskYesNo(fmt.Sprintf("Anda yakin ingin menghapus %d file?", len(selected)), false)
	if err != nil {
		return common.HandleInputError(err)
	}
	if !ok {
		ui.PrintInfo("Penghapusan dibatalkan oleh pengguna.")
		return nil
	}

	for _, p := range selected {
		if err := fs.RemoveFile(p); err != nil {
			s.Logger.Error(fmt.Sprintf("Gagal menghapus file %s: %v", p, err))
			ui.PrintError(fmt.Sprintf("Gagal menghapus file %s: %v", p, err))
		} else {
			s.Logger.Info(fmt.Sprintf("Berhasil menghapus: %s", p))
			ui.PrintSuccess(fmt.Sprintf("Berhasil menghapus: %s", p))
		}
	}

	return nil
}
