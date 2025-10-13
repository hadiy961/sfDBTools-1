package compress

import (
	"compress/gzip"
	"compress/zlib"
	"io"

	"github.com/klauspost/compress/zstd"
	"github.com/klauspost/pgzip"
)

// createGzipWriter creates a gzip writer with specified level
func createGzipWriter(w io.Writer, level CompressionLevel) (*gzip.Writer, error) {
	var gzipLevel int
	switch level {
	case LevelBestSpeed:
		gzipLevel = gzip.BestSpeed
	case LevelFast:
		gzipLevel = gzip.BestSpeed
	case LevelDefault:
		gzipLevel = gzip.DefaultCompression
	case LevelBetter:
		gzipLevel = gzip.BestCompression
	case LevelBest:
		gzipLevel = gzip.BestCompression
	default:
		gzipLevel = gzip.DefaultCompression
	}

	return gzip.NewWriterLevel(w, gzipLevel)
}

// createPgzipWriter creates a parallel gzip writer with specified level
func createPgzipWriter(w io.Writer, level CompressionLevel) (*pgzip.Writer, error) {
	var gzipLevel int
	switch level {
	case LevelBestSpeed:
		gzipLevel = pgzip.BestSpeed
	case LevelFast:
		gzipLevel = pgzip.BestSpeed
	case LevelDefault:
		gzipLevel = pgzip.DefaultCompression
	case LevelBetter:
		gzipLevel = pgzip.BestCompression
	case LevelBest:
		gzipLevel = pgzip.BestCompression
	default:
		gzipLevel = pgzip.DefaultCompression
	}

	return pgzip.NewWriterLevel(w, gzipLevel)
}

// createZlibWriter creates a zlib writer with specified level
func createZlibWriter(w io.Writer, level CompressionLevel) (*zlib.Writer, error) {
	var zlibLevel int
	switch level {
	case LevelBestSpeed:
		zlibLevel = zlib.BestSpeed
	case LevelFast:
		zlibLevel = zlib.BestSpeed
	case LevelDefault:
		zlibLevel = zlib.DefaultCompression
	case LevelBetter:
		zlibLevel = zlib.BestCompression
	case LevelBest:
		zlibLevel = zlib.BestCompression
	default:
		zlibLevel = zlib.DefaultCompression
	}

	return zlib.NewWriterLevel(w, zlibLevel)
}

// createZstdWriter creates a zstandard writer with specified level
func createZstdWriter(w io.Writer, level CompressionLevel) (*zstd.Encoder, error) {
	var zstdLevel zstd.EncoderLevel
	switch level {
	case LevelBestSpeed:
		zstdLevel = zstd.SpeedFastest
	case LevelFast:
		zstdLevel = zstd.SpeedDefault
	case LevelDefault:
		zstdLevel = zstd.SpeedDefault
	case LevelBetter:
		zstdLevel = zstd.SpeedBetterCompression
	case LevelBest:
		zstdLevel = zstd.SpeedBestCompression
	default:
		zstdLevel = zstd.SpeedDefault
	}

	return zstd.NewWriter(w, zstd.WithEncoderLevel(zstdLevel))
}
