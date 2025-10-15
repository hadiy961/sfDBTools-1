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
