// File : pkg/ui/ui_progressbar.go
// Deskripsi : Progress bar sederhana untuk terminal tanpa dependency eksternal
// Author : Copied/Adapted
// Tanggal : 2025-10-15
package ui

import (
	"fmt"
	"strings"
)

// ProgressBar is a minimal progress bar for terminal usage.
type ProgressBar struct {
	total   int
	current int
	width   int
	prefix  string
}

// NewProgressBar membuat instance baru progress bar.
func NewProgressBar(total int, prefix string) *ProgressBar {
	if total < 1 {
		total = 1
	}
	return &ProgressBar{total: total, current: 0, width: 40, prefix: prefix}
}

// Increment menambah progress satu unit dan menampilkan bar.
func (pb *ProgressBar) Increment(item string) {
	if pb.current < pb.total {
		pb.current++
	}
	percent := float64(pb.current) / float64(pb.total) * 100
	filled := int(float64(pb.width) * float64(pb.current) / float64(pb.total))
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", pb.width-filled)
	if item == "" {
		fmt.Printf("\r%s [%s] %3.0f%% (%d/%d)", pb.prefix, bar, percent, pb.current, pb.total)
	} else {
		fmt.Printf("\r%s [%s] %3.0f%% (%d/%d) %s", pb.prefix, bar, percent, pb.current, pb.total, item)
	}
	if pb.current >= pb.total {
		fmt.Println()
	}
}

// Finish menandakan progress selesai dan memaksa tampilan akhir.
func (pb *ProgressBar) Finish() {
	if pb.current < pb.total {
		pb.current = pb.total
	}
	pb.Increment("")
}
