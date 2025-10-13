package compress

import (
	"fmt"
	"strings"
)

// ValidateCompressionType validates if the compression type is supported
func ValidateCompressionType(compressionType string) (CompressionType, error) {
	ct := CompressionType(strings.ToLower(compressionType))
	switch ct {
	case CompressionNone, CompressionGzip, CompressionPgzip, CompressionZlib, CompressionZstd:
		return ct, nil
	default:
		return CompressionNone, fmt.Errorf("unsupported compression type: %s. Supported types: none, gzip, pgzip, zlib, zstd", compressionType)
	}
}

// ValidateCompressionLevel validates if the compression level is supported
func ValidateCompressionLevel(compressionLevel string) (CompressionLevel, error) {
	cl := CompressionLevel(strings.ToLower(compressionLevel))
	switch cl {
	case LevelBestSpeed, LevelFast, LevelDefault, LevelBetter, LevelBest:
		return cl, nil
	default:
		return LevelDefault, fmt.Errorf("unsupported compression level: %s. Supported levels: best_speed, fast, default, better, best", compressionLevel)
	}
}
