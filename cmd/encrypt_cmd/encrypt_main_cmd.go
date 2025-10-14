// File : cmd/encrypt_cmd/encrypt_main_cmd.go
// Deskripsi perintah utama 'encrypt' untuk enkripsi dan dekripsi file
// Author: Hadiyatna Muflihun
// Tanggal: 2024-10-14
// Last Modified: 2024-10-14

package encrypt_cmd

import (
	"github.com/spf13/cobra"
)

// EncryptCMD adalah perintah induk (parent command) untuk semua perintah 'encrypt'.
var EncryptCMD = &cobra.Command{
	Use:   "encrypt",
	Short: "Mengelola enkripsi dan dekripsi file (encrypt, decrypt)",
	Long: `Perintah 'encrypt' digunakan untuk mengelola proses enkripsi dan dekripsi file.
Terdapat beberapa sub-perintah seperti encrypt dan decrypt.
Gunakan 'encrypt <sub-command> --help' untuk informasi lebih lanjut tentang masing-masing sub-perintah.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	// Tambahkan sub-command ke parent command
	EncryptCMD.AddCommand(EncryptFileCmd)
	EncryptCMD.AddCommand(DecryptFileCmd)
}
