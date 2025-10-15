package backup

import (
	"errors"
	"sfDBTools/pkg/database"
	"time"
)

// SummaryFileEntry berisi informasi file summary untuk listing
type SummaryFileEntry struct {
	FileName      string
	FilePath      string
	BackupID      string
	Status        string
	BackupMode    string
	Timestamp     time.Time
	Duration      string
	DatabaseCount int
	FailedCount   int
	TotalSize     string
	CreatedAt     time.Time
}

var (
	// ErrGTIDUnsupported dikembalikan bila server tidak mendukung GTID (variabel tidak ada)
	ErrGTIDUnsupported = errors.New("gtid variables unsupported on server")
	// ErrGTIDPermissionDenied dikembalikan bila user tidak punya izin membaca GTID variables
	ErrGTIDPermissionDenied = errors.New("permission denied reading gtid variables")

	// ErrUserCancelled adalah sentinel error untuk menandai pembatalan oleh pengguna.
	ErrUserCancelled = errors.New("user_cancelled")

	// ErrNoDatabasesToBackup dikembalikan bila tidak ada database untuk di-backup setelah filtering
	ErrNoDatabasesToBackup = errors.New("tidak ada database untuk di-backup setelah filtering")
)

// DatabaseFilterStats menyimpan statistik hasil filtering database
type DatabaseFilterStats struct {
	TotalFound     int    // Total database yang ditemukan
	ToBackup       int    // Database yang akan di-backup
	ExcludedSystem int    // Database sistem yang dikecualikan
	ExcludedByList int    // Database dikecualikan karena blacklist
	ExcludedByFile int    // Database dikecualikan karena tidak ada di whitelist file
	ExcludedEmpty  int    // Database dengan nama kosong/invalid
	FilterMode     string // Mode filter: "whitelist", "blacklist", atau "system_only"
}

// BackupConfig untuk konfigurasi backup entry point
type BackupEntryConfig struct {
	HeaderTitle string
	ShowOptions bool
	BackupMode  string // "separate" atau "combined"
	EnableGTID  bool   // apakah perlu capture GTID
	SuccessMsg  string
	LogPrefix   string
}

// backupResult menampung semua informasi hasil dari proses backup.
type backupResult struct {
	successful []DatabaseBackupInfo
	failed     []FailedDatabaseInfo
	errors     []string
}

// jobResult adalah struct untuk mengirim hasil dari worker melalui channel.
type jobResult struct {
	info   DatabaseBackupInfo
	err    error
	dbName string // Diperlukan untuk logging error
}

// BackupSummary adalah struktur untuk menyimpan summary backup
type BackupSummary struct {
	// Informasi umum backup
	BackupID   string    `json:"backup_id"`
	Timestamp  time.Time `json:"timestamp"`
	BackupMode string    `json:"backup_mode"` // "separate" atau "combined"
	Status     string    `json:"status"`      // "success", "partial", "failed"
	Duration   string    `json:"duration"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`

	// Informasi database
	DatabaseStats DatabaseSummaryStats `json:"database_stats"`

	// Informasi file output
	OutputInfo OutputSummaryInfo `json:"output_info"`

	// Konfigurasi backup
	BackupConfig BackupConfigSummary `json:"backup_config"`

	// Database yang berhasil dan gagal
	SuccessfulDatabases []DatabaseBackupInfo `json:"successful_databases"`
	FailedDatabases     []FailedDatabaseInfo `json:"failed_databases"`

	// Detail database info
	DatabaseDetails map[string]database.DatabaseDetailInfo `json:"database_details,omitempty"` // Detail informasi setiap database

	// Informasi server database yang digunakan untuk backup
	ServerInfo ServerConnectionInfo `json:"server_info,omitempty"`

	// Informasi error (jika ada)
	Errors []string `json:"errors,omitempty"`
}

// DatabaseSummaryStats berisi statistik database
type DatabaseSummaryStats struct {
	TotalDatabases    int `json:"total_databases"`
	SuccessfulBackups int `json:"successful_backups"`
	FailedBackups     int `json:"failed_backups"`
	ExcludedDatabases int `json:"excluded_databases"`
	SystemDatabases   int `json:"system_databases"`
	FilteredDatabases int `json:"filtered_databases"`
}

// OutputSummaryInfo berisi informasi file output
type OutputSummaryInfo struct {
	OutputDirectory string            `json:"output_directory"`
	TotalFiles      int               `json:"total_files"`
	TotalSize       int64             `json:"total_size_bytes"`
	TotalSizeHuman  string            `json:"total_size_human"`
	Files           []SummaryFileInfo `json:"files"`
}

// BackupConfigSummary berisi ringkasan konfigurasi backup
type BackupConfigSummary struct {
	CompressionEnabled bool   `json:"compression_enabled"`
	CompressionType    string `json:"compression_type,omitempty"`
	CompressionLevel   string `json:"compression_level,omitempty"`
	EncryptionEnabled  bool   `json:"encryption_enabled"`
	DBListFile         string `json:"db_list_file,omitempty"`
	CleanupEnabled     bool   `json:"cleanup_enabled"`
	RetentionDays      int    `json:"retention_days,omitempty"`
}

// DatabaseBackupInfo berisi informasi database yang berhasil dibackup
type DatabaseBackupInfo struct {
	DatabaseName  string                       `json:"database_name"`
	OutputFile    string                       `json:"output_file"`
	FileSize      int64                        `json:"file_size_bytes"`
	FileSizeHuman string                       `json:"file_size_human"`
	Duration      string                       `json:"duration"`
	DetailInfo    *database.DatabaseDetailInfo `json:"detail_info,omitempty"` // Informasi detail database
}

// FailedDatabaseInfo berisi informasi database yang gagal dibackup
type FailedDatabaseInfo struct {
	DatabaseName string `json:"database_name"`
	Error        string `json:"error"`
}

// SummaryFileInfo berisi informasi file backup untuk summary (berbeda dari BackupFileInfo di cleanup)
type SummaryFileInfo struct {
	FileName     string    `json:"file_name"`
	FilePath     string    `json:"file_path"`
	Size         int64     `json:"size_bytes"`
	SizeHuman    string    `json:"size_human"`
	DatabaseName string    `json:"database_name,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// ServerConnectionInfo berisi informasi koneksi server (tanpa password)
type ServerConnectionInfo struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Database string `json:"database,omitempty"`
	Config   string `json:"config_name,omitempty"`
	Version  string `json:"version,omitempty"`
}
