// File : internal/backup/backup_helper.go
// Deskripsi : Helper functions untuk modul backup
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-08
// Last Modified : 2024-10-08
package backup

import (
	"context"
	"sfDBTools/pkg/database"
	"strings"

	"github.com/dustin/go-humanize"
)

// CheckDatabaseSize memeriksa ukuran database sebelum backup (database/database_count.go)
func (s *Service) CheckDatabaseSize(ctx context.Context, client *database.Client, dbName string) (int64, error) {
	size, err := client.GetDatabaseSize(ctx, dbName)
	if err != nil {
		s.Logger.Warn("Gagal mendapatkan ukuran database " + dbName + ": " + err.Error())
		return 0, err
	}

	//convert ukuran ke human readable pakai external package
	sizeHR := humanize.Bytes(uint64(size))

	// Log ukuran database
	s.Logger.Info("Ukuran database " + dbName + ": " + sizeHR)
	return size, nil
}

// sanitizeArgsForLogging menyembunyikan password dalam argumen untuk logging
func (s *Service) sanitizeArgsForLogging(args []string) []string {
	sanitized := make([]string, len(args))
	copy(sanitized, args)

	for i, arg := range sanitized {
		if strings.HasPrefix(arg, "--password=") {
			sanitized[i] = "--password=***"
		}
	}

	return sanitized
}
