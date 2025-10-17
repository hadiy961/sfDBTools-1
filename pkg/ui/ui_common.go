package ui

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
)

// formatDuration memformat durasi menjadi string human-readable
func FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%d jam %d menit %d detik", hours, minutes, seconds)
}

// formatFileSize memformat ukuran file menggunakan github.com/dustin/go-humanize
func FormatFileSize(size int64) string {
	return humanize.Bytes(uint64(size)) // contoh output: "12.3 MB"
}

// FormatBytes mengkonversi bytes (uint64) ke format human-readable menggunakan
// library github.com/dustin/go-humanize.
func FormatBytes(bytes uint64) string {
	// humanize.Bytes() menangani konversi menggunakan basis 1024 (KiB, MiB) secara otomatis
	return humanize.Bytes(bytes)
}

// FormatBytesInt64 mengkonversi int64 bytes ke format human-readable
func FormatBytesInt64(bytes int64) string {
	if bytes < 0 {
		return "0 B"
	}
	// Langsung konversi ke uint64 dan gunakan humanize.Bytes()
	return humanize.Bytes(uint64(bytes))
}
