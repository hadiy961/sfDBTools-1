// File : pkg/flag/dbconfig_flag.go
// Deskripsi : Fungsi utilitas untuk mendaftarkan flags pada perintah dbconfig
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03

package flags

import (
	"fmt"
	"os"
	"sfDBTools/internal/structs"

	"github.com/spf13/cobra"
)

// AddGenerateFlags adds flags specific to the generate command
func AddDBConfigCreateFlags(cmd *cobra.Command) {
	// 1. Siapkan struct target untuk pendaftaran flags
	flagStruct := &structs.DBConfigCreateFlags{}

	// 2. Daftarkan flags umum terlebih dahulu
	if err := DynamicAddFlags(cmd, flagStruct); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering AllDB flags dynamically: %v\n", err)
		os.Exit(1)
	}
}

// AddDBConfigEditFlags registers flags for the dbconfig edit command.
func AddDBConfigEditFlags(cmd *cobra.Command) {
	// Target struct untuk pendaftaran flags edit
	flagStruct := &structs.DBConfigEditFlags{}

	if err := DynamicAddFlags(cmd, flagStruct); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering DBConfig edit flags dynamically: %v\n", err)
		os.Exit(1)
	}
}

// AddDBConfigShowFlags registers flags for the dbconfig show command.
func AddDBConfigShowFlags(cmd *cobra.Command) {
	// Target struct untuk pendaftaran flags show
	flagStruct := &structs.DBConfigShowFlags{}

	if err := DynamicAddFlags(cmd, flagStruct); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering DBConfig show flags dynamically: %v\n", err)
		os.Exit(1)
	}
}
