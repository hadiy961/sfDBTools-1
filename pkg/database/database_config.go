package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	mysql "github.com/go-sql-driver/mysql"
)

// Config menyimpan parameter koneksi yang tidak akan berubah (immutable).
// Struct ini hanya bertanggung jawab untuk menyimpan data konfigurasi.
type Config struct {
	Host                 string
	Port                 int
	User                 string
	Password             string
	AllowNativePasswords bool
	ParseTime            bool
	Loc                  *time.Location
	Database             string // Optional, bisa kosong
}

// DSN menghasilkan string Data Source Name (DSN) dari konfigurasi.
func (c *Config) DSN() string {
	// Default ke time.Local jika c.Loc tidak di-set
	loc := time.Local
	if c.Loc != nil {
		loc = c.Loc
	}

	cfg := mysql.Config{
		User:                 c.User,
		Passwd:               c.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%d", c.Host, c.Port),
		DBName:               c.Database, // Set database name
		AllowNativePasswords: c.AllowNativePasswords,
		ParseTime:            c.ParseTime,
		Loc:                  loc,
	}
	return cfg.FormatDSN()
}

// Client adalah struct yang memegang koneksi database aktif (*sql.DB).
// Semua operasi ke database (Ping, Exec, Query) dilakukan melalui method dari struct ini.
type Client struct {
	db *sql.DB
}

// NewClient membuat instance Client baru, membuka koneksi pool, dan melakukan ping.
// Ini adalah satu-satunya tempat di mana sql.Open dipanggil.
func NewClient(ctx context.Context, cfg Config, timeout time.Duration, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration) (*Client, error) {
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi sql: %w", err)
	}

	// Atur parameter connection pool
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	// Gunakan context dengan timeout untuk ping awal
	pingCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		// Pastikan db ditutup jika ping gagal
		_ = db.Close()
		return nil, fmt.Errorf("gagal melakukan ping ke database: %w", err)
	}

	return &Client{db: db}, nil
}

// Close menutup connection pool. Wajib dipanggil saat aplikasi selesai.
func (c *Client) Close() error {
	return c.db.Close()
}

// Ping memeriksa apakah koneksi ke database masih hidup menggunakan pool yang ada.
func (c *Client) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// DB mengembalikan instance *sql.DB jika diperlukan akses langsung.
func (c *Client) DB() *sql.DB {
	return c.db
}

// GetMaxStatementsTime adalah helper internal untuk mengambil nilai @@max_statement_time.
func (c *Client) GetMaxStatementsTime(ctx context.Context) (float64, error) {
	var name string
	var raw sql.NullString
	if err := c.db.QueryRowContext(ctx, "SHOW GLOBAL VARIABLES LIKE 'max_statement_time'").Scan(&name, &raw); err != nil {
		// Jika tidak ada baris, MariaDB biasanya mengembalikan nilai default,
		// jadi error ini jarang terjadi kecuali ada masalah koneksi.
		return 0, fmt.Errorf("query max_statement_time gagal: %w", err)
	}

	if !raw.Valid || raw.String == "" {
		// Ini berarti nilainya tidak di-set atau 0
		return 0, nil
	}

	val, err := strconv.ParseFloat(raw.String, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing max_statement_time '%s' gagal: %w", raw.String, err)
	}
	return val, nil
}

// SetMaxStatementsTime mengatur nilai max_statements_time untuk sesi saat ini.
func (c *Client) SetMaxStatementsTime(ctx context.Context, seconds float64) error {
	_, err := c.db.ExecContext(ctx, "SET GLOBAL max_statement_time = ?", seconds)
	if err != nil {
		return fmt.Errorf("gagal set GLOBAL max_statement_time: %w", err)
	}
	return nil
}

// GetVersion mendapatkan versi server database sebagai string.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	var version string
	if err := c.db.QueryRowContext(ctx, "SELECT VERSION()").Scan(&version); err != nil {
		return "", fmt.Errorf("gagal mendapatkan versi database: %w", err)
	}
	return version, nil
}
