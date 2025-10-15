// File : internal/structs/structs_connection.go
// Deskripsi : Structs untuk koneksi database
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package structs

// ServerDBConnection - Database connection related flags
type ServerDBConnection struct {
	Host     string `flag:"host" env:"SFDB_DB_HOST" default:"localhost"`
	Port     int    `flag:"port" env:"SFDB_DB_PORT" default:"3306"`
	User     string `flag:"user" env:"SFDB_DB_USER" default:"root"`
	Password string `flag:"password" env:"SFDB_DB_PASSWORD" default:""`
	Database string
	Version  string // mariadb or mysql
}

// SourceDBConnection - Database source connection related flags
type SourceDBConnection struct {
	ServerDBConnection
	Database string `flag:"source-database" env:"SFDB_SOURCE_DB_NAME" default:"sourcedb"`
}

// DBConnection - Database connection related flags
type DBConnection struct {
	ServerDBConnection ServerDBConnection
	Database           string `flag:"database" env:"SFDB_DB_NAME" default:"testdb"`
	EncryptionKey      string `flag:"encryption-key" env:"SFDB_ENCRYPTION_KEY" default:"mysecretkey"`
}

// DBTestConnectionFlags - Flags for testing database connection
type DBTestConnectionFlags struct {
	ServerDBConnection ServerDBConnection
	TestDatabase       string `flag:"test-database" env:"SFDB_TEST_DB_NAME" default:""`
	CreateDatabase     bool   `flag:"create-database" env:"SFDB_CREATE_DATABASE" default:"false"`
	Verbose            bool   `flag:"verbose" env:"SFDB_VERBOSE" default:"false"`
}
