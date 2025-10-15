package database

import (
	"context"
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

// Database metric collection functions

func (s *Client) GetDatabaseSize(ctx context.Context, dbName string) (int64, error) {
	query := `
		SELECT COALESCE(SUM(data_length + index_length), 0) 
		FROM information_schema.tables 
		WHERE table_schema = ?`

	var size int64
	err := s.DB().QueryRowContext(ctx, query, dbName).Scan(&size)
	return size, err
}

func (s *Client) GetTableCount(ctx context.Context, dbName string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = ? AND table_type = 'BASE TABLE'`

	var count int
	err := s.DB().QueryRowContext(ctx, query, dbName).Scan(&count)
	return count, err
}

func (s *Client) GetProcedureCount(ctx context.Context, dbName string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.routines 
		WHERE routine_schema = ? AND routine_type = 'PROCEDURE'
		LIMIT 1000`

	var count int
	err := s.DB().QueryRowContext(ctx, query, dbName).Scan(&count)
	if err != nil {
		return 0, nil // Return 0 instead of error untuk graceful degradation
	}
	return count, err
}

func (s *Client) GetFunctionCount(ctx context.Context, dbName string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.routines 
		WHERE routine_schema = ? AND routine_type = 'FUNCTION'
		LIMIT 1000`

	var count int
	err := s.DB().QueryRowContext(ctx, query, dbName).Scan(&count)
	if err != nil {
		return 0, nil // Return 0 instead of error untuk graceful degradation
	}
	return count, err
}

func (s *Client) GetViewCount(ctx context.Context, dbName string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM information_schema.views 
		WHERE table_schema = ?
		LIMIT 1000`

	var count int
	err := s.DB().QueryRowContext(ctx, query, dbName).Scan(&count)
	if err != nil {
		return 0, nil // Return 0 instead of error untuk graceful degradation
	}
	return count, err
}

func (s *Client) GetUserGrantCount(ctx context.Context, dbName string) (int, error) {
	query := `
		SELECT COUNT(DISTINCT grantee) 
		FROM information_schema.schema_privileges 
		WHERE table_schema = ?
		LIMIT 1000`

	var count int
	err := s.DB().QueryRowContext(ctx, query, dbName).Scan(&count)
	if err != nil {
		return 0, nil // Return 0 instead of error untuk graceful degradation
	}
	return count, err
}
