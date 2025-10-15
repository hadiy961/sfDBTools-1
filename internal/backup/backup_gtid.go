// File : internal/backup/backup_gtid.go
// Deskripsi : GTID (Global Transaction Identifier) operations untuk backup
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-08
// Last Modified : 2024-10-15

package backup

import (
	"context"
	"sfDBTools/pkg/database"
)

// CaptureGTIDIfNeeded menangani pengambilan GTID jika opsi diaktifkan
func (s *Service) CaptureGTIDIfNeeded(ctx context.Context, client *database.Client) error {
	if s.BackupAll != nil && s.BackupAll.CaptureGtid {
		// Cek dukungan GTID
		enabled, pos, err := client.GetGTID(ctx)

		if err != nil {
			s.Logger.Warn("Gagal mendapatkan status GTID: " + err.Error())
			s.Logger.Warn("Menonaktifkan opsi capture GTID.")
			if s.BackupAll != nil {
				s.BackupAll.CaptureGtid = false
			}
			return nil
		}

		if enabled {
			s.Logger.Info("GTID saat ini: " + pos)
			// Simpan posisi GTID awal untuk referensi
			if s.BackupAll != nil {
				s.BackupAll.BackupInfo.GTIDCaptured = pos
			}
		} else {
			s.Logger.Warn("GTID tidak diaktifkan pada server ini.")
			s.Logger.Warn("Menonaktifkan opsi capture GTID.")
			if s.BackupAll != nil {
				s.BackupAll.CaptureGtid = false
			}
		}
	}
	return nil
}
