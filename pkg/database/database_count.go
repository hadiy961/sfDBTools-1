package database

import (
	"context"
	"database/sql"
)

// CountDatabases menghitung jumlah database pada server
func (s *Client) CountDatabases() (int, error) {
	var count int
	row := s.db.QueryRow("SELECT COUNT(*) FROM information_schema.SCHEMATA")
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// GetDatbaseSize mendapatkan ukuran database dalam bytes
func (s *Client) GetDatabaseSize(ctx context.Context, dbName string) (int64, error) {
	var size sql.NullInt64
	query := `
		SELECT SUM(data_length + index_length)
		FROM information_schema.TABLES
		WHERE table_schema = ?
		GROUP BY table_schema
	`
	err := s.db.QueryRowContext(ctx, query, dbName).Scan(&size)
	if err != nil {
		return 0, err
	}
	if !size.Valid {
		return 0, nil // Database kosong atau tidak ada tabel
	}
	return size.Int64, nil
}
