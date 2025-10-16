// File : cmd/dbscan_cmd/dbscan_all_cmd.go
// Deskripsi : Command untuk scan semua database
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025

package dbscan_cmd

import (
	"sfDBTools/internal/dbscan"
	defaultvalue "sfDBTools/internal/default_value"
	"sfDBTools/internal/structs"
	flags "sfDBTools/pkg/flag"
	"sfDBTools/pkg/globals"

	"github.com/spf13/cobra"
)

var scanAllOpts structs.ScanOptions

var ScanAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Scan semua database dan collect informasi detail",
	Long: `Scan semua database dari server yang dikonfigurasi dan mengumpulkan informasi detail.
	
Hasil scanning dapat disimpan ke database_details dan database_detail_history table untuk tracking dan monitoring.

Contoh penggunaan:
  sfdbtools dbscan all --config-file=/path/to/config.cnf
  sfdbtools dbscan all --config-file=/path/to/config.cnf --save-to-db=true
  sfdbtools dbscan all --config-file=/path/to/config.cnf --display-results=true
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := globals.GetLogger()
		config := globals.GetConfig()

		// Buat service
		svc := dbscan.NewService(logger, config)
		svc.SetScanOptions(scanAllOpts)

		// Execute scan
		scanConfig := dbscan.ScanEntryConfig{
			HeaderTitle: "Database Scanning - Semua Database",
			ShowOptions: true,
			SuccessMsg:  "Proses scanning database selesai.",
			LogPrefix:   "Proses database scan",
			Mode:        "all",
		}

		return svc.ExecuteScanCommand(scanConfig)
	},
}

func init() {
	// Set default values
	defaultOpts := defaultvalue.GetDefaultScanOptions()
	scanAllOpts = defaultOpts

	// Tambahkan flags menggunakan dynamic flag system
	flags.AddDbScanFlags(ScanAllCmd, &scanAllOpts)
}
