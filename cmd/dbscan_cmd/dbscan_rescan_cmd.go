// File : cmd/dbscan_cmd/dbscan_rescan_cmd.go
// Deskripsi : Command untuk re-scan database yang gagal
// Author : Hadiyatna Muflihun
// Tanggal : 16 Oktober 2025
// Last Modified : 16 Oktober 2025

package dbscan_cmd

import (
	"sfDBTools/internal/dbscan"
	defaultvalue "sfDBTools/internal/default_value"
	"sfDBTools/internal/structs"
	flags "sfDBTools/pkg/flag"
	"sfDBTools/pkg/globals"

	"github.com/spf13/cobra"
)

var scanRescanOpts structs.ScanOptions

var ScanRescanCmd = &cobra.Command{
	Use:   "rescan",
	Short: "Re-scan database yang gagal di-scan sebelumnya",
	Long: `Re-scan database yang gagal di-scan sebelumnya berdasarkan error message di database_details.

Command ini akan melakukan query ke table database_details untuk mencari database 
yang memiliki error_message IS NOT NULL, kemudian melakukan scan ulang untuk 
database-database tersebut.

Contoh penggunaan:
  sfdbtools dbscan rescan --config-file=/path/to/config.cnf
  sfdbtools dbscan rescan --config-file=/path/to/config.cnf --save-to-db=true
  sfdbtools dbscan rescan --config-file=/path/to/config.cnf --display-results=true
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := globals.GetLogger()
		config := globals.GetConfig()

		// Buat service
		svc := dbscan.NewService(logger, config)
		svc.SetScanOptions(scanRescanOpts)

		// Execute scan
		scanConfig := dbscan.ScanEntryConfig{
			HeaderTitle: "Database Scanning - Rescan Failed Databases",
			ShowOptions: true,
			SuccessMsg:  "Proses rescan database selesai.",
			LogPrefix:   "Proses database rescan",
			Mode:        "rescan",
		}

		return svc.ExecuteScanCommand(scanConfig)
	},
}

func init() {
	// Set default values
	defaultOpts := defaultvalue.GetDefaultScanOptions()
	scanRescanOpts = defaultOpts

	// Tambahkan flags menggunakan dynamic flag system
	flags.AddDbScanFlags(ScanRescanCmd, &scanRescanOpts)
}
