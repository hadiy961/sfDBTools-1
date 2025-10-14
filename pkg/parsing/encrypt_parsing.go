// File : pkg/parsing/encrypt_parsing.go
// Deskripsi : Fungsi utilitas untuk parsing flags encrypt dan decrypt
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-14
// Last Modified : 2024-10-14

package parsing

import (
	"sfDBTools/internal/structs"

	"github.com/spf13/cobra"
)

// ParseEncryptFlags mengurai flags dari command encrypt dan mengembalikan struct EncryptFlags
func ParseEncryptFlags(cmd *cobra.Command) (*structs.EncryptFlags, error) {
	flags := &structs.EncryptFlags{}

	// Parse flags menggunakan reflection
	if err := DynamicParseFlags(cmd, flags); err != nil {
		return nil, err
	}

	return flags, nil
}

// ParseDecryptFlags mengurai flags dari command decrypt dan mengembalikan struct DecryptFlags
func ParseDecryptFlags(cmd *cobra.Command) (*structs.DecryptFlags, error) {
	flags := &structs.DecryptFlags{}

	// Parse flags menggunakan reflection
	if err := DynamicParseFlags(cmd, flags); err != nil {
		return nil, err
	}

	return flags, nil
}

// ParseCleanupFlags mengurai flags dari command cleanup dan mengembalikan struct CleanupFlags
func ParseCleanupFlags(cmd *cobra.Command) (*structs.CleanupFlags, error) {
	flags := &structs.CleanupFlags{}

	// Parse flags menggunakan reflection
	if err := DynamicParseFlags(cmd, flags); err != nil {
		return nil, err
	}

	return flags, nil
}
