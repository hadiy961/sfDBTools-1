package structs

import "time"

// DBConfigCreateFlags - Flags for dbconfig create command
type DBConfigCreateFlags struct {
	DBConfigInfo DBConfigInfo
	Interactive  bool `flag:"interactive" env:"SFDB_INTERACTIVE" default:"true"`
}

// DBConfigInfo - Struct to hold database configuration information
type DBConfigInfo struct {
	ConfigName         string `flag:"config-name" env:"SFDB_CONFIG_NAME" default:"local_mariadb"`
	ServerDBConnection ServerDBConnection
	EncryptionKey      string `flag:"encryption-key" env:"SFDB_ENCRYPTION_KEY" default:""`
	FileSize           string
	LastModified       time.Time
	IsValid            bool
	FilePath           string
}

// DBConfigEditFlags - Flags for dbconfig edit command
type DBConfigEditFlags struct {
	File         string `flag:"file" env:"SFDB_CONFIG_FILE" default:""`
	Interactive  bool   `flag:"interactive" env:"SFDB_INTERACTIVE" default:"true"`
	DBConfigInfo DBConfigInfo
}

// DBConfigShowFlags - Flags for dbconfig show and validate commands
type DBConfigShowFlags struct {
	File           string `flag:"file" env:"SFDB_CONFIG_FILE" default:""`
	EncryptionKey  string `flag:"encryption-key" env:"SFDB_ENCRYPTION_KEY" default:""`
	RevealPassword bool   `flag:"reveal-password" env:"" default:"false"`
}
