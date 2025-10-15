// File : pkg/database/database_env_config.go
// Deskripsi : Fungsi untuk konfigurasi database menggunakan environment variables
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025
package database

import (
	"context"
	"fmt"
	"os"
	"sfDBTools/internal/structs"
	"strconv"
)

// InitializeDatabaseFromEnv membuat koneksi database menggunakan environment variables
// Environment variables yang digunakan:
// - SFDB_DB_HOST: Host database (default: localhost)
// - SFDB_DB_PORT: Port database (default: 3306)
// - SFDB_DB_USER: Username database (default: root)
// - SFDB_DB_PASSWORD: Password database (default: "")
// - SFDB_DB_NAME: Nama database (opsional, untuk pembuatan database)
func InitializeDatabaseFromEnv() (*Client, error) {
	// Ambil konfigurasi dari environment variables
	creds := structs.ServerDBConnection{
		Host:     GetEnvOrDefault("SFDB_DB_HOST", "localhost"),
		Port:     GetEnvOrDefaultInt("SFDB_DB_PORT", 3306),
		User:     GetEnvOrDefault("SFDB_DB_USER", "root"),
		Password: GetEnvOrDefault("SFDB_DB_PASSWORD", ""),
		Database: GetEnvOrDefault("SFDB_DB_NAME", ""),
	}

	// Menggunakan fungsi InitializeDatabase yang sudah ada
	return InitializeDatabase(creds)
}

// DatabaseExists mengecek apakah database dengan nama tertentu sudah ada
func (c *Client) DatabaseExists(ctx context.Context, dbName string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM information_schema.SCHEMATA WHERE SCHEMA_NAME = ?"
	err := c.db.QueryRowContext(ctx, query, dbName).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("gagal mengecek keberadaan database: %w", err)
	}
	return count > 0, nil
}

// Helper functions untuk mengambil environment variables
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
