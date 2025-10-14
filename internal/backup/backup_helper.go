// File : internal/backup/backup_helper.go
// Deskripsi : Helper functions untuk modul backup
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-08
// Last Modified : 2024-10-08
package backup

import (
	"strings"
)

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
