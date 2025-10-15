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

// DatabaseDetailInfo berisi informasi detail database
type DatabaseDetailInfo struct {
	DatabaseName   string `json:"database_name"`
	SizeBytes      int64  `json:"size_bytes"`
	SizeHuman      string `json:"size_human"`
	TableCount     int    `json:"table_count"`
	ProcedureCount int    `json:"procedure_count"`
	FunctionCount  int    `json:"function_count"`
	ViewCount      int    `json:"view_count"`
	UserGrantCount int    `json:"user_grant_count"`
	CollectionTime string `json:"collection_time"`
	Error          string `json:"error,omitempty"` // jika ada error saat collect
}

// DatabaseDetailJob untuk worker pattern
type DatabaseDetailJob struct {
	DatabaseName string
	Client       *database.Client
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
