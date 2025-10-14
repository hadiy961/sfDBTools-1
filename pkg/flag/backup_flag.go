// File : pkg/flag/backup_flag.go
// Deskripsi : Fungsi utilitas untuk mendaftarkan flags pada perintah backup
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-08
// Last Modified : 2024-10-08

package flags

import (
	"fmt"
	"os"
	defaultvalue "sfDBTools/internal/default_value"
	"sfDBTools/internal/structs"

	"github.com/spf13/cobra"
)

// AddBackupAllFlags adds flags specific to the backup command
// It will attempt to load default values from the application's default value
// provider so that Cobra help shows defaults coming from configuration.
func AddBackupAllFlags(cmd *cobra.Command) {
	// Try to get defaults from config; fall back to empty struct on error
	flagStruct, err := defaultvalue.GetDefaultBackupAllFlags()
	if err != nil {
		// If loading defaults fails, fall back to zero-value struct but warn user
		fmt.Fprintf(os.Stderr, "Warning: failed to load backup defaults: %v\n", err)
		flagStruct = &structs.BackupAllFlags{}
	}

	if err := DynamicAddFlags(cmd, flagStruct); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering AllDB flags dynamically: %v\n", err)
		os.Exit(1)
	}
}

// AddBackupDBFlags adds flags specific to the backup db command
// It will attempt to load default values from the application's default value
// provider so that Cobra help shows defaults coming from configuration.
func AddBackupDBFlags(cmd *cobra.Command) {
	// Try to get defaults from config; fall back to empty struct on error
	flagStruct, err := defaultvalue.GetDefaultBackupFlags()
	if err != nil {
		// If loading defaults fails, fall back to zero-value struct but warn user
		fmt.Fprintf(os.Stderr, "Warning: failed to load backup defaults: %v\n", err)
		flagStruct = &structs.BackupDBFlags{}
	}

	if err := DynamicAddFlags(cmd, flagStruct); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering DB flags dynamically: %v\n", err)
		os.Exit(1)
	}
}
