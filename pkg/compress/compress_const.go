package compress

import "io"

// CompressionType represents the type of compression algorithm
type CompressionType string

const (
	CompressionNone  CompressionType = "none"
	CompressionGzip  CompressionType = "gzip"
	CompressionPgzip CompressionType = "pgzip" // Parallel gzip
	CompressionZlib  CompressionType = "zlib"
	CompressionZstd  CompressionType = "zstd" // Zstandard
)

// CompressionLevel represents the compression level
type CompressionLevel string

const (
	LevelBestSpeed CompressionLevel = "best_speed"
	LevelFast      CompressionLevel = "fast"
	LevelDefault   CompressionLevel = "default"
	LevelBetter    CompressionLevel = "better"
	LevelBest      CompressionLevel = "best"
)

// CompressionConfig holds compression configuration
type CompressionConfig struct {
	Type  CompressionType
	Level CompressionLevel
}

// CompressingWriter wraps an io.Writer with compression
type CompressingWriter struct {
	baseWriter      io.Writer
	compressor      io.WriteCloser
	compressionType CompressionType
}
