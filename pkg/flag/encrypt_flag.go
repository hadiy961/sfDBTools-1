// File : pkg/flag/encrypt_flag.go
// Deskripsi : Fungsi utilitas untuk mendaftarkan flags pada perintah encrypt
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-14
// Last Modified : 2024-10-14

package flags

import (
	"fmt"
	"os"
	"sfDBTools/internal/structs"

	"github.com/spf13/cobra"
)

// AddEncryptFlags adds flags specific to the encrypt command
func AddEncryptFlags(cmd *cobra.Command) {
	// Gunakan struct kosong sebagai default
	flagStruct := &structs.EncryptFlags{}

	if err := DynamicAddFlags(cmd, flagStruct); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering Encrypt flags dynamically: %v\n", err)
		os.Exit(1)
	}
}

// AddDecryptFlags adds flags specific to the decrypt command
func AddDecryptFlags(cmd *cobra.Command) {
	// Gunakan struct kosong sebagai default
	flagStruct := &structs.DecryptFlags{}

	if err := DynamicAddFlags(cmd, flagStruct); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering Decrypt flags dynamically: %v\n", err)
		os.Exit(1)
	}
}

// AddCleanupFlags adds flags specific to the cleanup command
func AddCleanupFlags(cmd *cobra.Command) {
	// Gunakan struct kosong sebagai default
	flagStruct := &structs.CleanupFlags{}

	if err := DynamicAddFlags(cmd, flagStruct); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering Cleanup flags dynamically: %v\n", err)
		os.Exit(1)
	}
}
