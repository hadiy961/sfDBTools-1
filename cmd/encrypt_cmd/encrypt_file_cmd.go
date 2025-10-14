// File : cmd/encrypt_cmd/encrypt_file_cmd.go
// Deskripsi : Command untuk enkripsi file
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-14
// Last Modified : 2024-10-14

package encrypt_cmd

import (
	"fmt"
	"sfDBTools/internal/encrypt"
	flags "sfDBTools/pkg/flag"
	"sfDBTools/pkg/globals"
	"sfDBTools/pkg/parsing"

	"github.com/spf13/cobra"
)

// EncryptFileCmd adalah command untuk mengenkripsi file
var EncryptFileCmd = &cobra.Command{
	Use:   "file",
	Short: "Enkripsi file menggunakan AES-256-GCM",
	Long: `Mengenkripsi file menggunakan algoritma AES-256-GCM.
File hasil enkripsi kompatibel dengan format OpenSSL.

Contoh penggunaan:
  sfdbtools encrypt file --input backup.sql --output backup.sql.enc
  sfdbtools encrypt file --input backup.sql --key "mypassword"
  SFDB_ENCRYPTION_KEY="mypass" sfdbtools encrypt file --input backup.sql`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Validasi dependensi global
		if globals.Deps == nil || globals.Deps.Config == nil || globals.Deps.Logger == nil {
			return fmt.Errorf("dependency (config/logger) belum diinisialisasi")
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		logger := globals.GetLogger()
		cfg := globals.GetConfig()

		logger.Info("Memulai proses enkripsi file...")

		// Parse flags dari command
		encryptFlags, err := parsing.ParseEncryptFlags(cmd)
		if err != nil {
			logger.Errorf("Gagal mem-parse flags: %v", err)
			return err
		}

		// Debug: Log flags yang diterima
		logger.Debugf("Input file: %s", encryptFlags.InputFile)
		logger.Debugf("Output file: %s", encryptFlags.OutputFile)
		logger.Debugf("Key: %s", encryptFlags.EncryptionKey)

		// Buat service dengan flags yang sudah di-parse
		svc := encrypt.NewService(logger, cfg, encryptFlags)

		// Jalankan enkripsi
		if err := svc.EncryptFile(); err != nil {
			logger.Errorf("Enkripsi gagal: %v", err)
			return err
		}

		return nil
	},
}

func init() {
	// Daftarkan flags menggunakan fungsi khusus
	flags.AddEncryptFlags(EncryptFileCmd)
}
