package structs

import "time"

// DatabaseSizeInfo menyimpan informasi ukuran database
type DatabaseSizeInfo struct {
	DatabaseName string
	DataSize     int64
	IndexSize    int64
	TotalSize    int64
}

// DatabaseDetail menyimpan informasi detail database dari tabel database_details
type DatabaseDetail struct {
	DatabaseName   string    `db:"database_name"`
	SizeBytes      int64     `db:"size_bytes"`
	SizeHuman      string    `db:"size_human"`
	TableCount     int       `db:"table_count"`
	ProcedureCount int       `db:"procedure_count"`
	FunctionCount  int       `db:"function_count"`
	ViewCount      int       `db:"view_count"`
	UserGrantCount int       `db:"user_grant_count"`
	CollectionTime time.Time `db:"collection_time"`
	ErrorMessage   *string   `db:"error_message"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}
