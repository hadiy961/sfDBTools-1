package database

import (
	"context"
	"sfDBTools/internal/structs"
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
func (s *Client) GetDatabaseSize(ctx context.Context, dbName string) (*structs.DatabaseSizeInfo, error) {
	query := `
		SELECT 
			COALESCE(SUM(data_length), 0) as data_size,
			COALESCE(SUM(index_length), 0) as index_size,
			COALESCE(SUM(data_length + index_length), 0) as total_size
		FROM information_schema.tables 
		WHERE table_schema = ?
	`

	var dataSize, indexSize, totalSize int64
	err := s.db.QueryRowContext(ctx, query, dbName).Scan(&dataSize, &indexSize, &totalSize)
	if err != nil {
		return nil, err
	}

	return &structs.DatabaseSizeInfo{
		DatabaseName: dbName,
		DataSize:     dataSize,
		IndexSize:    indexSize,
		TotalSize:    totalSize,
	}, nil
}
