// File : internal/structs/structs_backup.go
// Deskripsi : Struct untuk menyimpan flags pada perintah backup
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-08
// Last Modified : 2024-10-08

package structs

// CompressionOptions - Opsi kompresi untuk backup
type CompressionOptions struct {
	Type    string `flag:"compress-type" env:"SFDB_COMPRESSION_TYPE" default:"gzip"`   // Jenis kompresi
	Level   string `flag:"compress-level" env:"SFDB_COMPRESSION_LEVEL" default:"best"` // Level kompresi, jika berlaku
	Enabled bool   `flag:"compress" env:"SFDB_COMPRESSION_ENABLED" default:"true"`     // Apakah kompresi diaktifkan
}

// BackupOptions - Struct untuk menyimpan opsi pada perintah backup
type BackupOptions struct {
	Encryption      EncryptionOptions
	Compression     CompressionOptions
	OutputDirectory string `flag:"output" env:"SFDB_BACKUP_OUTPUT_DIR" default:"./backups"`
	OutputFile      string // Nama file output spesifik (jika kosong, gunakan format default)
	DBConfig        DBConfigInfo
	DiskCheck       bool `flag:"disk-check" env:"SFDB_VERIFICATION_DISK_CHECK" default:"true"` // Apakah cek disk diaktifkan
	Exclude         ExcludeOptions
	DBList          string `flag:"db-list" env:"SFDB_BACKUP_DB_LIST_FILE" default:""`
	Cleanup         CleanupOptions
}

// EncryptionOptions - Opsi enkripsi untuk backup
type EncryptionOptions struct {
	Enabled bool   `flag:"encrypt" env:"SFDB_ENCRYPTION_ENABLED"` // Apakah enkripsi diaktifkan
	Key     string `flag:"encrypt-key" env:"SFDB_ENCRYPTION_KEY"` // Kunci enkripsi
}

// CleanupOptions - Opsi pembersihan backup lama
type CleanupOptions struct {
	Enabled       bool   `flag:"cleanup" env:"SFDB_CLEANUP_ENABLED" default:"false"` // Aktifkan pembersihan otomatis backup lama sesuai kebijakan retensi
	Scheduled     string // Jadwal pembersihan: daily, weekly, monthly
	RetentionDays int    `flag:"cleanup-days" env:"SFDB_CLEANUP_DAYS" default:"30"` // Jumlah hari untuk menyimpan backup sebelum dihapus (retensi)
}

// CleanupFlags - Flags khusus untuk command cleanup (lebih sederhana)
type CleanupFlags struct {
	OutputDirectory string `flag:"output" env:"SFDB_BACKUP_OUTPUT_DIR" default:"/mnt/nfs/backup"` // Direktori backup yang akan dibersihkan
	RetentionDays   int    `flag:"cleanup-days" env:"SFDB_CLEANUP_DAYS" default:"30"`             // Jumlah hari untuk menyimpan backup sebelum dihapus
	Enabled         bool   `flag:"cleanup" env:"SFDB_CLEANUP_ENABLED" default:"true"`             // Aktifkan pembersihan (harus true untuk menjalankan cleanup)
	DryRun          bool   `flag:"dry-run" env:"SFDB_CLEANUP_DRY_RUN" default:"false"`            // Mode dry-run: tampilkan file yang akan dihapus tanpa menghapus
	Pattern         string `flag:"pattern" env:"SFDB_CLEANUP_PATTERN" default:""`                 // Pattern file yang akan dibersihkan (opsional)
}

// BackupAllFlags - Struct untuk menyimpan flags pada perintah backup
type BackupAllFlags struct {
	BackupOptions BackupOptions
	BackupInfo    BackupInfo
	CaptureGtid   bool   `flag:"capture-gtid" env:"SFDB_CAPTURE_GTID"`         // Apakah GTID capture diaktifkan
	Mode          string `flag:"mode" env:"SFDB_BACKUP_MODE" default:"single"` // Mode backup: single atau multi
	// Cache internal untuk optimasi performa
	DbListCache map[string]bool // Cache untuk database whitelist dari file
}

// BackupDBFlags - Struct untuk menyimpan flags pada perintah backup db
type BackupDBFlags struct {
	BackupOptions BackupOptions
	BackupInfo    BackupInfo
	DBName        []string `flag:"db" env:"SFDB_BACKUP_DB_NAME" default:""` // Nama database yang akan dibackup
}

// DBListOptions - Struct untuk menyimpan flags pada perintah backup db-list
type DBListOptions struct {
	DBList string `flag:"db-list" env:"SFDB_DB_LIST_FILE" default:""`
}

// ExcludeOptions - Struct untuk menyimpan flags pada perintah backup exclude
type ExcludeOptions struct {
	Databases []string `flag:"exclude-db" env:"SFDB_BACKUP_EXCLUDE_DATABASES" default:""` // Daftar database yang dikecualikan, dipisah koma
	// Tables    []string `flag:"exclude-tables" env:"SFDB_BACKUP_EXCLUDE_TABLES" default:""`       // Daftar tabel yang dikecualikan, dipisah koma
	SystemsDB bool `flag:"exclude-system" env:"SFDB_BACKUP_EXCLUDE_SYSTEMS" default:"false"` // Apakah sistem dikecualikan
	Users     bool `flag:"exclude-user" env:"SFDB_BACKUP_EXCLUDE_USERS" default:"false"`     // Apakah user dikecualikan
	Data      bool `flag:"exclude-data" env:"SFDB_BACKUP_EXCLUDE_DATA" default:"false"`      // Apakah exclude data (hanya struktur)
}

// BackupInfo - Struct untuk menyimpan informasi hasil backup
type BackupInfo struct {
	Enabled       bool   // Apakah pembuatan info backup diaktifkan
	FilePath      string // Path lengkap file backup yang dihasilkan
	FileSize      int64  // Ukuran file backup dalam bytes
	Compression   string // Jenis kompresi yang digunakan (jika ada)
	Encryption    bool   // Apakah file dienkripsi
	DatabaseCount int    // Jumlah database yang dibackup
	TablesCount   int    // Jumlah tabel yang dibackup
	DataOnly      bool   // Apakah hanya data yang dibackup (tanpa struktur)
	Duration      int64  // Durasi proses backup dalam detik
	GTIDCaptured  string // Posisi GTID yang ditangkap (jika ada)
}

type FilterInfo struct {
	TotalDatabases    int
	ExcludedDatabases int
	IncludedDatabases int
	SystemDatabases   int
}

// BackupSummaryFlags - Struct untuk menyimpan flags pada perintah backup summary
type BackupSummaryFlags struct {
	BackupID string `flag:"backup-id" env:"SFDB_BACKUP_SUMMARY_ID" default:""`       // ID backup untuk ditampilkan
	Latest   bool   `flag:"latest" env:"SFDB_BACKUP_SUMMARY_LATEST" default:"false"` // Tampilkan summary terbaru
}
