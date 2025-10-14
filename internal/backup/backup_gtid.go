// File : internal/backup/backup_gtid.go
// Deskripsi : GTID (Global Transaction Identifier) operations untuk backup
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-08
// Last Modified : 2024-10-15

package backup

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
	"strconv"
	"strings"
)

// GetGTID memeriksa apakah GTID diaktifkan dan mengembalikan posisinya jika ada.
func (s *Service) GetGTID(ctx context.Context, client *database.Client) (bool, string, error) {
	ui.PrintSubHeader("Memeriksa Status GTID")

	var domainID sql.NullInt64
	row := client.DB().QueryRowContext(ctx, "SELECT @@GLOBAL.gtid_domain_id")
	if err := row.Scan(&domainID); err != nil {
		// QueryRow.Scan mengembalikan error bila variabel tidak ada, bila akses ditolak, dll.
		le := strings.ToLower(err.Error())
		if err == sql.ErrNoRows {
			s.Logger.Debug("Query GTID returned no rows")
			return false, "", nil
		}
		// Deteksi kasus umum dan kembalikan error sentinel agar caller dapat menanggapinya spesifik
		if strings.Contains(le, "unknown system variable") || strings.Contains(le, "gtid_domain_id") {
			s.Logger.Debug("Server tidak mendukung GTID: " + err.Error())
			return false, "", ErrGTIDUnsupported
		}
		if strings.Contains(le, "access denied") || strings.Contains(le, "permission denied") {
			s.Logger.Debug("Tidak ada izin membaca GTID variables: " + err.Error())
			return false, "", ErrGTIDPermissionDenied
		}
		// Unexpected error -> wrap dan return
		return false, "", fmt.Errorf("gagal membaca gtid_domain_id: %w", err)
	}

	if !domainID.Valid {
		s.Logger.Debug("MariaDB GTID domain_id NULL atau tidak di-set.")
		return false, "", nil
	}
	s.Logger.Info("MariaDB GTID domain_id: " + strconv.FormatInt(domainID.Int64, 10))

	// Cek juga gtid_current_pos
	var currentPos sql.NullString
	row = client.DB().QueryRowContext(ctx, "SELECT @@GLOBAL.gtid_current_pos")
	if err := row.Scan(&currentPos); err != nil {
		le := strings.ToLower(err.Error())
		if err == sql.ErrNoRows {
			s.Logger.Debug("Query gtid_current_pos returned no rows")
			return true, "", nil
		}
		if strings.Contains(le, "unknown system variable") || strings.Contains(le, "gtid_current_pos") {
			s.Logger.Debug("Server tidak mendukung gtid_current_pos: " + err.Error())
			return true, "", ErrGTIDUnsupported
		}
		if strings.Contains(le, "access denied") || strings.Contains(le, "permission denied") {
			s.Logger.Debug("Tidak ada izin membaca gtid_current_pos: " + err.Error())
			return true, "", ErrGTIDPermissionDenied
		}
		return true, "", fmt.Errorf("gagal membaca gtid_current_pos: %w", err)
	}

	if !currentPos.Valid || currentPos.String == "" {
		s.Logger.Debug("gtid_current_pos kosong.")
		return false, "", nil
	}

	return true, currentPos.String, nil
}

// CaptureGTIDIfNeeded menangani pengambilan GTID jika opsi diaktifkan
func (s *Service) CaptureGTIDIfNeeded(ctx context.Context, client *database.Client) error {
	if s.BackupAll != nil && s.BackupAll.CaptureGtid {
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
			if s.BackupAll != nil {
				s.BackupAll.CaptureGtid = false
			}
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
