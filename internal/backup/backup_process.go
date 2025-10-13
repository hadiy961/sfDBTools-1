package backup

import (
	"context"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
)

// DumpAllDB adalah fungsi internal untuk melakukan backup semua database
func (s *Service) DumpAllDB(ctx context.Context, client *database.Client, mode string) error {
	ui.PrintSubHeader("Proses Backup")

	// Panggil fungsi backup semua database
	s.Logger.Info("Memulai proses backup database...")

	// 5. mysqldump semu database sekaligus
	s.Logger.Info("Proses backup semua database selesai.")

	return nil
}
