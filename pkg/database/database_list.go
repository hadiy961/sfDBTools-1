package database

import (
	"context"
	"errors"
)

// getDatabaseList mendapatkan daftar database dari server, menerapkan filter exclude jika ada.
func (s *Client) GetDatabaseList(ctx context.Context, client *Client) ([]string, error) {
	var databases []string

	rows, err := s.db.QueryContext(ctx, "SHOW DATABASES")
	if err != nil {
		return nil, errors.New("gagal mendapatkan daftar database: " + err.Error())
	}
	defer rows.Close()

	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			return nil, errors.New("gagal membaca nama database: " + err.Error())
		}

		databases = append(databases, dbName)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.New("terjadi kesalahan saat membaca daftar database: " + err.Error())
	}

	return databases, nil
}
