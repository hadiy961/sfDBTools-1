package backup

import (
	"context"
	"errors"
	"sfDBTools/pkg/database"
)

// CaptureGTIDIfNeeded menangani pengambilan GTID jika opsi diaktifkan
func (s *Service) CaptureGTIDIfNeeded(ctx context.Context, client *database.Client) error {
	if s.BackupAll.CaptureGtid {
		// Cek dukungan GTID
		enabled, pos, err := s.GetGTID(ctx, client)

		if err != nil {
			if errors.Is(err, ErrGTIDUnsupported) {
				s.Logger.Warn("Server tidak mendukung GTID: " + err.Error())
			} else if errors.Is(err, ErrGTIDPermissionDenied) {
				s.Logger.Warn("Tidak memiliki izin untuk membaca variabel GTID: " + err.Error())
			} else {
				s.Logger.Warn("Gagal memeriksa dukungan GTID: " + err.Error())
			}
			s.Logger.Warn("Menonaktifkan opsi capture GTID.")
			s.BackupAll.CaptureGtid = false
		}

		if enabled {
			s.Logger.Info("GTID saat ini: " + pos)
			// Simpan posisi GTID awal untuk referensi
			s.BackupAll.BackupInfo.GTIDCaptured = pos
		} else {
			s.Logger.Warn("GTID tidak diaktifkan pada server ini.")
			s.Logger.Warn("Menonaktifkan opsi capture GTID.")
			s.BackupAll.CaptureGtid = false
		}
	}
	return nil
}
