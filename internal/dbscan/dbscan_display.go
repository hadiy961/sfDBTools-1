package dbscan

import (
	"fmt"
	"sfDBTools/pkg/ui"
)

// DisplayScanOptions menampilkan opsi scanning yang sedang aktif.
func (s *Service) DisplayScanOptions() {
	ui.PrintSubHeader("Opsi Scanning")
	targetConn := s.getTargetDBConfig()

	data := [][]string{
		{"Exclude System DB", fmt.Sprintf("%v", s.ScanOptions.ExcludeSystem)},
		{"Include List", fmt.Sprintf("%d database", len(s.ScanOptions.IncludeList))},
		{"Exclude List", fmt.Sprintf("%d database", len(s.ScanOptions.ExcludeList))},
		{"Save to DB", fmt.Sprintf("%v", s.ScanOptions.SaveToDB)},
		{"Display Results", fmt.Sprintf("%v", s.ScanOptions.DisplayResults)},
	}

	if s.ScanOptions.SaveToDB {
		targetInfo := fmt.Sprintf("%s@%s:%d/%s",
			targetConn.User, targetConn.Host, targetConn.Port, targetConn.Database)
		data = append(data, []string{"Target DB", targetInfo})
	}

	ui.FormatTable([]string{"Parameter", "Value"}, data)
}

// DisplayFilterStats menampilkan statistik hasil pemfilteran database.
func (s *Service) DisplayFilterStats(stats *DatabaseFilterStats) {
	ui.PrintSubHeader("Statistik Filtering Database")
	data := [][]string{
		{"Total Ditemukan", fmt.Sprintf("%d", stats.TotalFound)},
		{"Akan di-scan", ui.ColorText(fmt.Sprintf("%d", stats.ToScan), ui.ColorGreen)},
		{"Dikecualikan (Sistem)", fmt.Sprintf("%d", stats.ExcludedSystem)},
		{"Dikecualikan (Exclude List)", fmt.Sprintf("%d", stats.ExcludedByList)},
		{"Dikecualikan (Bukan di Include List)", fmt.Sprintf("%d", stats.ExcludedByFile)},
		{"Dikecualikan (Nama Kosong)", fmt.Sprintf("%d", stats.ExcludedEmpty)},
	}
	ui.FormatTable([]string{"Kategori", "Jumlah"}, data)
}