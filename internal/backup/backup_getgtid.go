package backup

import (
	"context"
	"database/sql"
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
