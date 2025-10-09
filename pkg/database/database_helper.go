package database

import (
	"context"
	"fmt"
	"sfDBTools/internal/structs"
	"time"
)

func InitializeDatabase(creds structs.ServerDBConnection) (*Client, error) {
	// Konfigurasi dasar yang umum digunakan
	cfg := Config{
		Host:                 creds.Host,
		Port:                 creds.Port,
		User:                 creds.User,
		Password:             creds.Password,
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
