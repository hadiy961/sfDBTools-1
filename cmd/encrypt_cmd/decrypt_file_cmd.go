// File : cmd/encrypt_cmd/decrypt_file_cmd.go
// Deskripsi : Command untuk dekripsi file
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

// DecryptFileCmd adalah command untuk mendekripsi file
var DecryptFileCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Dekripsi file yang dienkripsi dengan AES-256-GCM",
	Long: `Mendekripsi file yang dienkripsi menggunakan algoritma AES-256-GCM.
Mendukung file yang dienkripsi dengan format OpenSSL.

Contoh penggunaan:
  sfdbtools encrypt decrypt --input backup.sql.enc --output backup.sql
  sfdbtools encrypt decrypt --input backup.sql.enc --key "mypassword"
  SFDB_ENCRYPTION_KEY="mypass" sfdbtools encrypt decrypt --input backup.sql.enc`,

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

		logger.Info("Memulai proses dekripsi file...")

		// Parse flags dari command
		decryptFlags, err := parsing.ParseDecryptFlags(cmd)
		if err != nil {
			logger.Errorf("Gagal mem-parse flags: %v", err)
			return err
		}

		// Debug: Log flags yang diterima
		logger.Debugf("Input file: %s", decryptFlags.InputFile)
		logger.Debugf("Output file: %s", decryptFlags.OutputFile)
		logger.Debugf("Key: %s", decryptFlags.EncryptionKey)

		// Buat service dengan flags yang sudah di-parse
		svc := encrypt.NewService(logger, cfg, decryptFlags)

		// Jalankan dekripsi
		if err := svc.DecryptFile(); err != nil {
			logger.Errorf("Dekripsi gagal: %v", err)
			return err
		}

		return nil
	},
}

func init() {
	// Daftarkan flags menggunakan fungsi khusus
	flags.AddDecryptFlags(DecryptFileCmd)
}
