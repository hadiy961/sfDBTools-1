package ui

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

func RunWithSpinner(message, successMessage string, task func() error) error {
	// Buat dan konfigurasi spinner
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " " + message
	s.Start()

	// Jalankan tugas yang diberikan
	err := task()

	// Hentikan spinner
	s.Stop()

	// Cek hasil dan tampilkan pesan
	if err != nil {
		// \r (carriage return) memindahkan kursor ke awal baris
		// untuk menimpa teks spinner.
		fmt.Printf("\r❌ Gagal: %v\n", err)
	} else {
		fmt.Printf("\r✅ %s\n", successMessage)
	}

	// Kembalikan error agar bisa ditangani oleh pemanggil fungsi jika perlu
	return err
}
