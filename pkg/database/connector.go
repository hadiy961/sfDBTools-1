package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	mysql "github.com/go-sql-driver/mysql"
)

// PingMySQL mencoba melakukan ping ke server MySQL dengan timeout tertentu.
// Jika berhasil, mengembalikan nil. Jika gagal, mengembalikan error.
func PingMySQL(ctx context.Context, host string, port int, user, password string, timeout time.Duration) error {
	// Konfigurasi DSN MySQL
	cfg := mysql.Config{
		User:                 user,
		Passwd:               password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", host, port),
		AllowNativePasswords: true,
	}

	// Buka koneksi database
	dsn := cfg.FormatDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	// Set batas waktu koneksi
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return db.PingContext(cctx)
}
