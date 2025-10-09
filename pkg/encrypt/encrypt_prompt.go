// File : pkg/encrypt/encrypt_prompt.go
// Deskripsi : Fungsi utilitas untuk prompt input user pada modul enkripsi
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03

package encrypt

import (
	"fmt"
	"os"
	"sfDBTools/pkg/common"
	"sfDBTools/pkg/input"
	"sfDBTools/pkg/ui"
	"strings"

	"github.com/AlecAivazis/survey/v2"
)

// EncryptionPrompt mendapatkan password enkripsi dari env atau prompt.
// Mengembalikan password, sumbernya ("env" atau "prompt"), dan error jika ada.
func EncryptionPrompt(promptMessage string) (string, string, error) {
	// Cek environment variable SFDB_ENCRYPTION_KEY
	// Jika ada, gunakan itu
	// Jika tidak, minta user memasukkan password
	if password := os.Getenv(common.SFDB_ENCRYPTION_KEY); password != "" {
		return password, "env", nil
	} else {
		ui.PrintSubHeader("Authentication Required")
		ui.PrintWarning("Environment variable SFDB_ENCRYPTION_KEY tidak ditemukan atau kosong. Silakan atur SFDB_ENCRYPTION_KEY atau ketik password.")
	}
	// Minta user memasukkan password
	// Validator: tidak boleh kosong
	// return input.AskPassword(promptMessage, survey.Required)
	EncryptionPassword, err := input.AskPassword("Encryption Password", survey.Required)

	return EncryptionPassword, "prompt", err

}

// GetEncryptionPassword mendapatkan password enkripsi dari env atau prompt.
func GetEncryptionPassword() (string, string, error) {
	// Dapatkan password enkripsi dari env atau prompt
	encryptionPassword, source, err := EncryptionPrompt("🔑 Encryption password: ")
	if err != nil {
		return "", "", fmt.Errorf("failed to get encryption password: %w", err)
	}

	// Validasi password enkripsi
	// if len(encryptionPassword) < 8 {
	// 	return "", source, fmt.Errorf("encryption password must be at least 8 characters long")
	// }
	// Jika valid, kembalikan password dan sumbernya
	return encryptionPassword, source, nil
}

// resolveEncryptionKey mengembalikan kunci enkripsi final dan sumbernya.
// Urutan: existing (flag/state) -> env/prompt (via encrypt.GetEncryptionPassword).
func ResolveEncryptionKey(existing string) (string, string, error) {
	if k := strings.TrimSpace(existing); k != "" {
		return k, "flag/state", nil
	}
	// Jika tidak ada existing, minta dari env atau prompt
	pwd, source, err := GetEncryptionPassword()
	if err != nil {
		return "", source, err
	}
	// Validasi tambahan: pastikan tidak kosong setelah trim
	pwd = strings.TrimSpace(pwd)
	if pwd == "" {
		return "", source, fmt.Errorf("kunci enkripsi kosong")
	}
	return pwd, source, nil
}
