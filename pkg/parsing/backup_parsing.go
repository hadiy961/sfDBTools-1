// File : pkg/parsing/backup_parsing.go
// Deskripsi : Fungsi utilitas untuk parsing flags perintah backup
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-08
// Last Modified : 2024-10-08

package parsing

import (
	"fmt"
	defaultvalue "sfDBTools/internal/default_value"
	"sfDBTools/internal/structs"

	"github.com/spf13/cobra"
)

func ParseBackupAllFlags(cmd *cobra.Command) (GenerateDefault *structs.BackupAllFlags, err error) {
	// 1. Dapatkan nilai default dari Configuration.
	// Nilai ini akan menjadi fallback terakhir.
	GenerateDefault, err = defaultvalue.GetDefaultBackupAllFlags()
	if err != nil {
		// Kembalikan error agar caller (seperti fungsi Run Cobra) yang menanganinya.
		return nil, fmt.Errorf("failed to load general backup defaults from config: %w", err)
	}

	// 2. Parse flags dinamis ke dalam struct menggunakan refleksi.
	// Flags sudah didaftarkan pada init/command setup, jadi kita hanya membaca
	// nilainya ke dalam struct target.
	if err := DynamicParseFlags(cmd, GenerateDefault); err != nil {
		return nil, fmt.Errorf("failed to dynamically parse general backup flags: %w", err)
	}

	// 3. Kembalikan struct yang sudah diisi.
	return GenerateDefault, nil
}
