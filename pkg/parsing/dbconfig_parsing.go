package parsing

import (
	"fmt"
	defaultvalue "sfDBTools/internal/default_value"
	"sfDBTools/internal/structs"

	"github.com/spf13/cobra"
)

func ParseDBConfigCreateFlags(cmd *cobra.Command) (GenerateDefault *structs.DBConfigCreateFlags, err error) {
	// 1. Dapatkan nilai default dari Configuration.
	// Nilai ini akan menjadi fallback terakhir.
	GenerateDefault, err = defaultvalue.GetDefaultDBConfigCreate()
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

// ParseDBConfigEditFlags
func ParseDBConfigEditFlags(cmd *cobra.Command) (GenerateDefault *structs.DBConfigEditFlags, err error) {
	// 1. Dapatkan nilai default dari Configuration.
	// Nilai ini akan menjadi fallback terakhir.
	GenerateDefault, err = defaultvalue.GetDefaultDBConfigEdit()
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
