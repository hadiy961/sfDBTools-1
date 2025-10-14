// File : pkg/encrypt/encrypt_file.go
// Deskripsi : Fungsi utilitas untuk enkripsi dan dekripsi file backup
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-14
// Last Modified : 2024-10-14
package encrypt

import (
	"fmt"
	"io"
	"os"
)

// EncryptFile mengenkripsi file menggunakan passphrase
func EncryptFile(inputPath, outputPath string, passphrase []byte) error {
	// Buka file input
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("gagal membuka file input: %w", err)
	}
	defer inputFile.Close()

	// Buat file output
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("gagal membuat file output: %w", err)
	}
	defer outputFile.Close()

	// Buat encrypting writer
	encryptingWriter, err := NewEncryptingWriter(outputFile, passphrase)
	if err != nil {
		return fmt.Errorf("gagal membuat encrypting writer: %w", err)
	}
	defer encryptingWriter.Close()

	// Copy file dengan enkripsi
	_, err = io.Copy(encryptingWriter, inputFile)
	if err != nil {
		return fmt.Errorf("gagal mengenkripsi file: %w", err)
	}

	return nil
}

// DecryptFile mendekripsi file menggunakan passphrase
func DecryptFile(inputPath, outputPath string, passphrase []byte) error {
	// Baca file terenkripsi
	encryptedData, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("gagal membaca file terenkripsi: %w", err)
	}

	// Dekripsi data
	plaintext, err := DecryptAES(encryptedData, passphrase)
	if err != nil {
		return fmt.Errorf("gagal mendekripsi file: %w", err)
	}

	// Tulis hasil dekripsi
	err = os.WriteFile(outputPath, plaintext, 0644)
	if err != nil {
		return fmt.Errorf("gagal menulis file hasil dekripsi: %w", err)
	}

	return nil
}

// IsEncryptedFile memeriksa apakah file adalah file terenkripsi berdasarkan header
func IsEncryptedFile(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("gagal membuka file: %w", err)
	}
	defer file.Close()

	// Baca 8 byte pertama untuk cek header "Salted__"
	header := make([]byte, 8)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		return false, fmt.Errorf("gagal membaca header file: %w", err)
	}

	if n < 8 {
		return false, nil // File terlalu kecil untuk menjadi file terenkripsi
	}

	opensslHeader := []byte("Salted__")
	for i := 0; i < 8; i++ {
		if header[i] != opensslHeader[i] {
			return false, nil
		}
	}

	return true, nil
}
