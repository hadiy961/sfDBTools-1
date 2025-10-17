// File : cmd/dbscan_cmd/dbscan_main_cmd.go
// Deskripsi : Command utama untuk database scan
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025

package dbscan_cmd

import (
	"fmt"
	"sfDBTools/pkg/globals"

	"github.com/spf13/cobra"
)

var DbScanCmd = &cobra.Command{
	Use:   "dbscan",
	Short: "Scan dan collect informasi detail database",
	Long: `Command dbscan digunakan untuk melakukan scanning database dan mengumpulkan informasi detail seperti:
- Ukuran database (bytes dan human-readable)
- Jumlah tabel, stored procedure, function, view
- User grant count

Hasil scanning dapat disimpan ke database untuk tracking dan monitoring.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Validasi global dependencies
		logger := globals.GetLogger()
		config := globals.GetConfig()
		if logger == nil || config == nil {
			return fmt.Errorf("dependency (config/logger) belum diinisialisasi")
		}
		return nil
	},
}

func init() {
	// Tambahkan sub-commands
	DbScanCmd.AddCommand(ScanAllCmd)
	DbScanCmd.AddCommand(ScanDatabaseCmd)
	DbScanCmd.AddCommand(ScanRescanCmd)
	DbScanCmd.AddCommand(ScanSingleDBCmd)
}
