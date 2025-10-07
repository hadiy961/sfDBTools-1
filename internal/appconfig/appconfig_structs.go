// File : internal/appconfig/appconfig_structs.go
// Deskripsi : Struct untuk konfigurasi aplikasi yang di-load dari file YAML
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package appconfig

// Config adalah struktur level atas yang memegang semua bagian konfigurasi.
// Tag 'yaml' digunakan untuk memetakan field Go ke kunci di file YAML.
type Config struct {
	Backup      BackupConfig      `yaml:"backup"`
	ConfigDir   ConfigDirConfig   `yaml:"config_dir"`
	General     GeneralConfig     `yaml:"general"`
	Log         LogConfig         `yaml:"log"`
	Mariadb     MariadbConfig     `yaml:"mariadb"`
	SystemUsers SystemUsersConfig `yaml:"system_users"`
}

// Struct untuk bagian 'backup'
type BackupConfig struct {
	Compression   CompressionConfig  `yaml:"compression"`
	MysqlDumpArgs string             `yaml:"mysqldump_args"`
	Retention     RetentionConfig    `yaml:"retention"`
	Security      SecurityConfig     `yaml:"security"`
	Output        OutputConfig       `yaml:"output"`
	Verification  VerificationConfig `yaml:"verification"`
}

type CompressionConfig struct {
	Algorithm string `yaml:"algorithm"`
	Level     string `yaml:"level"`
	Required  bool   `yaml:"required"`
}

type RetentionConfig struct {
	CleanupEnabled  bool   `yaml:"cleanup_enabled"`
	CleanupSchedule string `yaml:"cleanup_schedule"`
	Days            int    `yaml:"days"`
}

type SecurityConfig struct {
	ChecksumVerification bool `yaml:"checksum_verification"`
	EncryptionRequired   bool `yaml:"encryption_required"`
	IntegrityCheck       bool `yaml:"integrity_check"`
}

type OutputConfig struct {
	BaseDirectory string `yaml:"base_directory"`
	CleanupTemp   bool   `yaml:"cleanup_temp"`
	Naming        struct {
		IncludeClientCode bool   `yaml:"include_client_code"`
		IncludeHostname   bool   `yaml:"include_hostname"`
		Pattern           string `yaml:"pattern"`
	} `yaml:"naming"`
	Structure struct {
		CreateSubdirs bool   `yaml:"create_subdirs"`
		Pattern       string `yaml:"pattern"`
	} `yaml:"structure"`
	TempDirectory string `yaml:"temp_directory"`
}

type VerificationConfig struct {
	CompareChecksums bool   `yaml:"compare_checksums"`
	DiskSpaceCheck   bool   `yaml:"disk_space_check"`
	MinimumFreeSpace string `yaml:"minimum_free_space"`
	VerifyAfterWrite bool   `yaml:"verify_after_write"`
}

// Struct untuk bagian 'config_dir'
type ConfigDirConfig struct {
	DatabaseConfig         string `yaml:"database_config"`
	DatabaseList           string `yaml:"database_list"`
	MariadbConfigTemplates string `yaml:"mariadb_config_templates"`
}

// Struct untuk bagian 'general'
type GeneralConfig struct {
	AppName    string `yaml:"app_name"`
	Author     string `yaml:"author"`
	ClientCode string `yaml:"client_code"`
	Locale     struct {
		DateFormat string `yaml:"date_format"`
		TimeFormat string `yaml:"time_format"`
		Timezone   string `yaml:"timezone"`
	} `yaml:"locale"`
	Version string `yaml:"version"`
}

// Struct untuk bagian 'log'
type LogConfig struct {
	Format string `yaml:"format"`
	Level  string `yaml:"level"`
	Output struct {
		Console struct {
			Enabled bool `yaml:"enabled"`
		} `yaml:"console"`
		File struct {
			Dir             string `yaml:"dir"`
			Enabled         bool   `yaml:"enabled"`
			FilenamePattern string `yaml:"filename_pattern"`
			Rotation        struct {
				CompressOld   bool   `yaml:"compress_old"`
				Daily         bool   `yaml:"daily"`
				MaxSize       string `yaml:"max_size"`
				RetentionDays int    `yaml:"retention_days"` // Menggunakan time.Duration untuk konversi
			} `yaml:"rotation"`
		} `yaml:"file"`
	} `yaml:"output"`
	Timezone string `yaml:"timezone"`
}

// Struct untuk bagian 'mariadb'
type MariadbConfig struct {
	BinlogDir           string `yaml:"binlog_dir"`
	ConfigDir           string `yaml:"config_dir"`
	DataDir             string `yaml:"data_dir"`
	EncryptionKeyFile   string `yaml:"encryption_key_file"`
	InnodbEncryptTables bool   `yaml:"innodb_encrypt_tables"`
	LogDir              string `yaml:"log_dir"`
	Port                int    `yaml:"port"`
	ServerID            int    `yaml:"server_id"`
	Version             string `yaml:"version"`
}

// Struct untuk bagian 'system_users'
type SystemUsersConfig struct {
	Users []string `yaml:"users"`
}
