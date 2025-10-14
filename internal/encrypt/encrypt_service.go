// File : internal/encrypt/encrypt_service.go
// Deskripsi : Implementasi service untuk encrypt dan decrypt file
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-14
// Last Modified : 2024-10-14

package encrypt

import (
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/ui"
	"strings"
)

// EncryptFile mengenkripsi file berdasarkan flags yang diberikan
func (s *Service) EncryptFile() error {
	if s.EncryptInfo == nil {
		return fmt.Errorf("encrypt flags tidak tersedia")
	}

	// Validasi input file
	if s.EncryptInfo.InputFile == "" {
		return fmt.Errorf("input file harus dispecifikasi")
	}

	// Cek apakah input file exists
	if _, err := os.Stat(s.EncryptInfo.InputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file tidak ditemukan: %s", s.EncryptInfo.InputFile)
	}

	// Generate output file jika tidak dispecifikasi
	outputFile := s.EncryptInfo.OutputFile
	if outputFile == "" {
		outputFile = s.EncryptInfo.InputFile + ".enc"
	}

	// Cek apakah output file sudah exists dan overwrite flag
	if _, err := os.Stat(outputFile); err == nil && !s.EncryptInfo.Overwrite {
		return fmt.Errorf("output file sudah exists: %s (gunakan --overwrite untuk menimpa)", outputFile)
	}

	// Resolve encryption key
	encryptionKey := s.EncryptInfo.EncryptionKey
	if encryptionKey == "" {
		resolvedKey, source, err := encrypt.ResolveEncryptionKey("")
		if err != nil {
			return fmt.Errorf("gagal mendapatkan kunci enkripsi: %w", err)
		}
		encryptionKey = resolvedKey
		s.Logger.Infof("Kunci enkripsi diperoleh dari: %s", source)
	}

	// Pastikan output directory exists
	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori output: %w", err)
	}

	ui.PrintSubHeader("Enkripsi File")
	s.Logger.Infof("Input file: %s", s.EncryptInfo.InputFile)
	s.Logger.Infof("Output file: %s", outputFile)

	// Enkripsi file
	if err := encrypt.EncryptFile(s.EncryptInfo.InputFile, outputFile, []byte(encryptionKey)); err != nil {
		return fmt.Errorf("gagal mengenkripsi file: %w", err)
	}

	ui.PrintSuccess(fmt.Sprintf("File berhasil dienkripsi: %s", outputFile))
	s.Logger.Info("Proses enkripsi selesai")

	return nil
}

// DecryptFile mendekripsi file berdasarkan flags yang diberikan
func (s *Service) DecryptFile() error {
	if s.DecryptInfo == nil {
		return fmt.Errorf("decrypt flags tidak tersedia")
	}

	// Validasi input file
	if s.DecryptInfo.InputFile == "" {
		return fmt.Errorf("input file harus dispecifikasi")
	}

	// Cek apakah input file exists
	if _, err := os.Stat(s.DecryptInfo.InputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file tidak ditemukan: %s", s.DecryptInfo.InputFile)
	}

	// Validasi apakah file terenkripsi
	isEncrypted, err := encrypt.IsEncryptedFile(s.DecryptInfo.InputFile)
	if err != nil {
		return fmt.Errorf("gagal memeriksa file enkripsi: %w", err)
	}
	if !isEncrypted {
		return fmt.Errorf("file input bukan file terenkripsi yang valid")
	}

	// Generate output file jika tidak dispecifikasi
	outputFile := s.DecryptInfo.OutputFile
	if outputFile == "" {
		// Hapus ekstensi .enc jika ada
		if strings.HasSuffix(s.DecryptInfo.InputFile, ".enc") {
			outputFile = strings.TrimSuffix(s.DecryptInfo.InputFile, ".enc")
		} else {
			outputFile = s.DecryptInfo.InputFile + ".decrypted"
		}
	}

	// Cek apakah output file sudah exists dan overwrite flag
	if _, err := os.Stat(outputFile); err == nil && !s.DecryptInfo.Overwrite {
		return fmt.Errorf("output file sudah exists: %s (gunakan --overwrite untuk menimpa)", outputFile)
	}

	// Resolve encryption key
	encryptionKey := s.DecryptInfo.EncryptionKey
	if encryptionKey == "" {
		resolvedKey, source, err := encrypt.ResolveEncryptionKey("")
		if err != nil {
			return fmt.Errorf("gagal mendapatkan kunci enkripsi: %w", err)
		}
		encryptionKey = resolvedKey
		s.Logger.Infof("Kunci enkripsi diperoleh dari: %s", source)
	}

	// Pastikan output directory exists
	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori output: %w", err)
	}

	ui.PrintSubHeader("Dekripsi File")
	s.Logger.Infof("Input file: %s", s.DecryptInfo.InputFile)
	s.Logger.Infof("Output file: %s", outputFile)

	// Dekripsi file
	if err := encrypt.DecryptFile(s.DecryptInfo.InputFile, outputFile, []byte(encryptionKey)); err != nil {
		return fmt.Errorf("gagal mendekripsi file: %w", err)
	}

	ui.PrintSuccess(fmt.Sprintf("File berhasil didekripsi: %s", outputFile))
	s.Logger.Info("Proses dekripsi selesai")

	return nil
}
