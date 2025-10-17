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

var scanDatabaseOpts structs.ScanOptions

var ScanDatabaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Scan database tertentu berdasarkan filter",
	Long: `Scan database tertentu berdasarkan include/exclude list atau database list file.

Contoh penggunaan:
  sfdbtools dbscan database --config-file=/path/to/config.cnf --db-list-file=db_list.txt
  sfdbtools dbscan database --config-file=/path/to/config.cnf --include=db1,db2,db3
  sfdbtools dbscan database --config-file=/path/to/config.cnf --exclude=test_db
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := globals.GetLogger()
		config := globals.GetConfig()

		// Buat service
		svc := dbscan.NewService(logger, config)
		svc.SetScanOptions(scanDatabaseOpts)

		// Execute scan
		scanConfig := dbscan.ScanEntryConfig{
			HeaderTitle: "Database Scanning - Database Tertentu",
			ShowOptions: true,
			SuccessMsg:  "Proses scanning database selesai.",
			LogPrefix:   "Proses database scan",
			Mode:        "database",
		}

		return svc.ExecuteScanCommand(scanConfig)
	},
}

func init() {
	// Set default values
	defaultOpts := defaultvalue.GetDefaultScanOptions("database")
	scanDatabaseOpts = defaultOpts

	// Tambahkan flags menggunakan dynamic flag system
	flags.AddDbScanFlags(ScanDatabaseCmd, &scanDatabaseOpts)
}
