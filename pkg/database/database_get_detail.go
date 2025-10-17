// File: pkg/database/database_get_detail.go
// Deskripsi: Fungsi untuk membaca detail database dari tabel database_details
// Author: Hadiyatna Muflihun
// Tanggal: 16 Oktober 2025
// Last Modified: 16 Oktober 2025

package database

import (
	"context"
	"database/sql"
	"fmt"
	"sfDBTools/internal/structs"
)

// GetDatabaseDetails mengambil detail untuk multiple database dari tabel database_details
// berdasarkan database_name, server_host, dan server_port
func (c *Client) GetDatabaseDetails(ctx context.Context, databaseNames []string, serverHost string, serverPort int) (map[string]structs.DatabaseDetail, error) {
	details := make(map[string]structs.DatabaseDetail)

	for _, dbName := range databaseNames {
		detail, err := c.GetSingleDatabaseDetail(ctx, dbName, serverHost, serverPort)
		if err != nil {
			// Skip database yang tidak ditemukan, lanjutkan ke database berikutnya
			if err.Error() == fmt.Sprintf("detail database tidak ditemukan untuk database '%s' di server %s:%d", dbName, serverHost, serverPort) {
				continue
			}
			return nil, fmt.Errorf("gagal mengambil detail untuk database '%s': %w", dbName, err)
		}
		if detail == nil {
			// Jika detailnya nil, lanjutkan ke database berikutnya
			continue
		}

		// Dereference pointer dan simpan sebagai value
		details[dbName] = *detail
	}

	return details, nil
}

// GetSingleDatabaseDetail mengambil detail database dari tabel database_details
// berdasarkan database_name, server_host, dan server_port
func (c *Client) GetSingleDatabaseDetail(ctx context.Context, databaseName, serverHost string, serverPort int) (*structs.DatabaseDetail, error) {
	query := `
		SELECT 
			database_name,
			size_bytes,
			size_human,
			table_count,
			procedure_count,
			function_count,
			view_count,
			user_grant_count,
			collection_time,
			error_message,
			created_at,
			updated_at
		FROM database_details
		WHERE database_name = ? 
			AND server_host = ? 
			AND server_port = ?
		ORDER BY collection_time DESC
		LIMIT 1
	`

	var detail structs.DatabaseDetail
	var errorMessage sql.NullString

	err := c.db.QueryRowContext(ctx, query, databaseName, serverHost, serverPort).Scan(
		&detail.DatabaseName,
		&detail.SizeBytes,
		&detail.SizeHuman,
		&detail.TableCount,
		&detail.ProcedureCount,
		&detail.FunctionCount,
		&detail.ViewCount,
		&detail.UserGrantCount,
		&detail.CollectionTime,
		&errorMessage,
		&detail.CreatedAt,
		&detail.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("detail database tidak ditemukan untuk database '%s' di server %s:%d", databaseName, serverHost, serverPort)
		}
		return nil, fmt.Errorf("gagal mengambil detail database: %w", err)
	}

	// Handle nullable error_message
	if errorMessage.Valid {
		detail.ErrorMessage = &errorMessage.String
	}

	return &detail, nil
}

// GetAllDatabaseDetails mengambil semua detail database dari tabel database_details
// untuk server tertentu
func (c *Client) GetAllDatabaseDetails(ctx context.Context, serverHost string, serverPort int) ([]structs.DatabaseDetail, error) {
	query := `
		SELECT 
			database_name,
			size_bytes,
			size_human,
			table_count,
			procedure_count,
			function_count,
			view_count,
			user_grant_count,
			collection_time,
			error_message,
			created_at,
			updated_at
		FROM database_details
		WHERE server_host = ? 
			AND server_port = ?
		ORDER BY database_name, collection_time DESC
	`

	rows, err := c.db.QueryContext(ctx, query, serverHost, serverPort)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar detail database: %w", err)
	}
	defer rows.Close()

	var details []structs.DatabaseDetail
	for rows.Next() {
		var detail structs.DatabaseDetail
		var errorMessage sql.NullString

		err := rows.Scan(
			&detail.DatabaseName,
			&detail.SizeBytes,
			&detail.SizeHuman,
			&detail.TableCount,
			&detail.ProcedureCount,
			&detail.FunctionCount,
			&detail.ViewCount,
			&detail.UserGrantCount,
			&detail.CollectionTime,
			&errorMessage,
			&detail.CreatedAt,
			&detail.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("gagal scan baris detail database: %w", err)
		}

		// Handle nullable error_message
		if errorMessage.Valid {
			detail.ErrorMessage = &errorMessage.String
		}

		details = append(details, detail)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error saat iterasi rows: %w", err)
	}

	return details, nil
}

// GetLatestDatabaseDetails mengambil detail database terbaru (berdasarkan collection_time)
// untuk semua database di server tertentu
func (c *Client) GetLatestDatabaseDetails(ctx context.Context, serverHost string, serverPort int) ([]structs.DatabaseDetail, error) {
	query := `
		SELECT 
			dd1.database_name,
			dd1.size_bytes,
			dd1.size_human,
			dd1.table_count,
			dd1.procedure_count,
			dd1.function_count,
			dd1.view_count,
			dd1.user_grant_count,
			dd1.collection_time,
			dd1.error_message,
			dd1.created_at,
			dd1.updated_at
		FROM database_details dd1
		INNER JOIN (
			SELECT database_name, MAX(collection_time) AS max_collection_time
			FROM database_details
			WHERE server_host = ? AND server_port = ?
			GROUP BY database_name
		) dd2 ON dd1.database_name = dd2.database_name 
			AND dd1.collection_time = dd2.max_collection_time
			AND dd1.server_host = ?
			AND dd1.server_port = ?
		ORDER BY dd1.database_name
	`

	rows, err := c.db.QueryContext(ctx, query, serverHost, serverPort, serverHost, serverPort)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil detail database terbaru: %w", err)
	}
	defer rows.Close()

	var details []structs.DatabaseDetail
	for rows.Next() {
		var detail structs.DatabaseDetail
		var errorMessage sql.NullString

		err := rows.Scan(
			&detail.DatabaseName,
			&detail.SizeBytes,
			&detail.SizeHuman,
			&detail.TableCount,
			&detail.ProcedureCount,
			&detail.FunctionCount,
			&detail.ViewCount,
			&detail.UserGrantCount,
			&detail.CollectionTime,
			&errorMessage,
			&detail.CreatedAt,
			&detail.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("gagal scan baris detail database: %w", err)
		}

		// Handle nullable error_message
		if errorMessage.Valid {
			detail.ErrorMessage = &errorMessage.String
		}

		details = append(details, detail)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error saat iterasi rows: %w", err)
	}

	return details, nil
}
