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
	DBConfig        DBConfigInfo
}

// EncryptionOptions - Opsi enkripsi untuk backup
type EncryptionOptions struct {
	Enabled bool   `flag:"encrypt" env:"SFDB_ENCRYPTION_ENABLED"` // Apakah enkripsi diaktifkan
	Key     string `flag:"encrypt-key" env:"SFDB_ENCRYPTION_KEY"` // Kunci enkripsi
}

// CleanupOptions - Opsi pembersihan backup lama
type CleanupOptions struct {
	Enabled       bool   `flag:"cleanup" env:"SFDB_CLEANUP_ENABLED" default:"false"`            // Aktifkan pembersihan otomatis backup lama sesuai kebijakan retensi
	Scheduled     string `flag:"cleanup-schedule" env:"SFDB_CLEANUP_SCHEDULED" default:"daily"` // Jadwal pembersihan: daily, weekly, monthly
	RetentionDays int    `flag:"cleanup-days" env:"SFDB_CLEANUP_DAYS" default:"30"`             // Jumlah hari untuk menyimpan backup sebelum dihapus (retensi)
}

// BackupAllFlags - Struct untuk menyimpan flags pada perintah backup
type BackupAllFlags struct {
	BackupOptions    BackupOptions
	DBList           DBListOptions
	Cleanup          CleanupOptions
	Exclude          ExcludeOptions
	Verification     VerificationOptions
	BackupInfo       BackupInfo
	CaptureGtid      bool `flag:"capture-gtid" env:"SFDB_CAPTURE_GTID"`             // Apakah GTID capture diaktifkan
	CreateBackupInfo bool `flag:"create-backup-info" env:"SFDB_CREATE_BACKUP_INFO"` // Internal: apakah membuat BackupInfo setelah backup selesai
}

// DBListOptions - Struct untuk menyimpan flags pada perintah backup db-list
type DBListOptions struct {
	File string `flag:"db-list" env:"SFDB_BACKUP_DB_LIST_FILE" default:""`
}

// ExcludeOptions - Struct untuk menyimpan flags pada perintah backup exclude
type ExcludeOptions struct {
	Databases []string `flag:"exclude-db" env:"SFDB_BACKUP_EXCLUDE_DATABASES" default:""`        // Daftar database yang dikecualikan, dipisah koma
	Tables    []string `flag:"exclude-tables" env:"SFDB_BACKUP_EXCLUDE_TABLES" default:""`       // Daftar tabel yang dikecualikan, dipisah koma
	SystemsDB bool     `flag:"exclude-system" env:"SFDB_BACKUP_EXCLUDE_SYSTEMS" default:"false"` // Apakah sistem dikecualikan
	Users     bool     `flag:"exclude-user" env:"SFDB_BACKUP_EXCLUDE_USERS" default:"false"`     // Apakah user dikecualikan
	Data      bool     `flag:"exclude-data" env:"SFDB_BACKUP_EXCLUDE_DATA" default:"false"`      // Apakah exclude data (hanya struktur)
}

// VerificationOptions - Opsi verifikasi backup
type VerificationOptions struct {
	DiskCheck bool `flag:"disk-check" env:"SFDB_VERIFICATION_DISK_CHECK" default:"true"` // Apakah cek disk diaktifkan
}

// BackupInfo - Struct untuk menyimpan informasi hasil backup
type BackupInfo struct {
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
