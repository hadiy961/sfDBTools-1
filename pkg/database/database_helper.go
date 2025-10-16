package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/ui"
	"strings"
	"time"
)

func InitializeDatabase(creds structs.ServerDBConnection) (*Client, error) {
	// Konfigurasi dasar yang umum digunakan
	cfg := Config{
		Host:                 creds.Host,
		Port:                 creds.Port,
		User:                 creds.User,
		Password:             creds.Password,
		Database:             creds.Database, // Include database name
		AllowNativePasswords: true,
		ParseTime:            true,
		Loc:                  time.Local,
	}

	// Konfigurasi connection pool (bisa diubah di sini jika perlu)
	const (
		connectTimeout  = 5 * time.Second
		maxOpenConns    = 10
		maxIdleConns    = 5
		connMaxLifetime = 5 * time.Minute
	)

	ctx := context.Background()

	// Membuat klien baru dengan semua konfigurasi di atas
	client, err := NewClient(ctx, cfg, connectTimeout, maxOpenConns, maxIdleConns, connMaxLifetime)
	if err != nil {
		// Kita bungkus errornya untuk memberikan konteks tambahan
		return nil, fmt.Errorf("gagal saat inisialisasi koneksi database: %w", err)
	}

	return client, nil
}

func (s *Client) GetGTID(ctx context.Context) (bool, string, error) {
	ui.PrintSubHeader("Memeriksa Status GTID")

	var domainID sql.NullInt64
	row := s.DB().QueryRowContext(ctx, "SELECT @@GLOBAL.gtid_domain_id")
	if err := row.Scan(&domainID); err != nil {
		// QueryRow.Scan mengembalikan error bila variabel tidak ada, bila akses ditolak, dll.
		le := strings.ToLower(err.Error())
		if err == sql.ErrNoRows {
			return false, "", errors.New("query gtid_domain_id returned no rows")
		}
		// Deteksi kasus umum dan kembalikan error sentinel agar caller dapat menanggapinya spesifik
		if strings.Contains(le, "unknown system variable") || strings.Contains(le, "gtid_domain_id") {
			return false, "", errors.New("gtid_domain_id tidak dikenali")
		}
		if strings.Contains(le, "access denied") || strings.Contains(le, "permission denied") {
			return false, "", errors.New("akses ditolak untuk membaca gtid_domain_id")
		}
		// Unexpected error -> wrap dan return
		return false, "", errors.New("gagal membaca gtid_domain_id: " + err.Error())
	}

	if !domainID.Valid {
		return false, "", errors.New("gtid_domain_id NULL atau tidak di-set")
	}

	// Cek juga gtid_current_pos
	var currentPos sql.NullString
	row = s.DB().QueryRowContext(ctx, "SELECT @@GLOBAL.gtid_current_pos")
	if err := row.Scan(&currentPos); err != nil {
		le := strings.ToLower(err.Error())
		if err == sql.ErrNoRows {
			return true, "", errors.New("query gtid_current_pos returned no rows")
		}
		if strings.Contains(le, "unknown system variable") || strings.Contains(le, "gtid_current_pos") {
			return true, "", errors.New("server tidak mendukung gtid_current_pos: " + err.Error())
		}
		if strings.Contains(le, "access denied") || strings.Contains(le, "permission denied") {
			return true, "", errors.New("tidak ada izin membaca gtid_current_pos: " + err.Error())
		}
		return true, "", errors.New("gagal membaca gtid_current_pos: " + err.Error())
	}

	if !currentPos.Valid || currentPos.String == "" {
		return false, "", errors.New("gtid_current_pos NULL atau kosong")
	}

	return true, currentPos.String, nil
}
