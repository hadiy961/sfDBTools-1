// File : cmd/dbscan_cmd/dbscan_database_cmd.go
// Deskripsi : Command untuk scan database tertentu
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

var scanSingleOpts structs.ScanOptions

var ScanSingleDBCmd = &cobra.Command{
	Use:   "single",
	Short: "Scan satu database tertentu",
	Long: `Scan satu database tertentu menggunakan --source-database.

Contoh penggunaan:
  sfdbtools dbscan single --config-file=/path/to/config.cnf --source-database=mydb
  sfdbtools dbscan single --config-file=/path/to/config.cnf --source-database=production_db --save-to-db
  sfdbtools dbscan single --source-database=test_db --target-host=localhost --target-database=sfdbtools
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := globals.GetLogger()
		config := globals.GetConfig()

		// Buat service
		svc := dbscan.NewService(logger, config)
		svc.SetScanOptions(scanSingleOpts)

		// Validasi SourceDatabase sudah diisi
		if scanSingleOpts.SourceDatabase == "" {
			return cmd.Usage()
		}

		// Execute scan
		scanConfig := dbscan.ScanEntryConfig{
			HeaderTitle: "Database Scanning - Single Database",
			ShowOptions: true,
			SuccessMsg:  "Proses scanning database selesai.",
			LogPrefix:   "Proses single database scan",
			Mode:        "single",
		}

		return svc.ExecuteScanCommand(scanConfig)
	},
}

func init() {
	// Set default values
	defaultOpts := defaultvalue.GetDefaultScanOptions("single")
	scanSingleOpts = defaultOpts

	// Tambahkan flags menggunakan dynamic flag system
	flags.AddDbScanFlags(ScanSingleDBCmd, &scanSingleOpts)
}
