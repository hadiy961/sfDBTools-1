package structs

import "sfDBTools/pkg/compress"

// --- Structs & Options ---

// EstimateOptions menampung semua parameter untuk proses estimasi.
type EstimateOptions struct {
	CompressionEnabled bool
	CompressionType    compress.CompressionType
	CompressionLevel   compress.CompressionLevel
	EncryptionEnabled  bool
	BackupMode         string
	SafetyMarginPct    float64
}

// BackupSizeEstimate menyimpan detail hasil estimasi untuk satu database.
type BackupSizeEstimate struct {
	DatabaseName         string
	OriginalSize         int64
	EstimatedSQLDumpSize uint64
	EstimatedFinalSize   uint64
	CompressionRatio     float64
	CompressionEnabled   bool
	EncryptionEnabled    bool
}

// DiskSpaceCheckResult menyimpan hasil akhir pengecekan ruang disk.
type DiskSpaceCheckResult struct {
	OutputDirectory         string
	EstimatedBackupSize     uint64
	RequiredWithMargin      uint64
	AvailableDiskSpace      uint64
	HasEnoughSpace          bool
	DatabaseEstimates       []BackupSizeEstimate
	DatabasesWithoutDetails int
}
