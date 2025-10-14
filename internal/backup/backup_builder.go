package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/encrypt"
	"strings"
)

// executeMysqldumpWithPipe menjalankan mysqldump dengan pipe ke file output
func (s *Service) executeMysqldumpWithPipe(ctx context.Context, mysqldumpArgs []string, outputPath string, compressionRequired bool, compressionType string) error {
	// Buat file output
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("gagal membuat file output: %w", err)
	}
	defer outputFile.Close()

	var writer io.Writer = outputFile
	var closers []io.Closer

	// Setup enkripsi jika diperlukan (layer pertama: paling dalam)
	// Urutan layer: File -> Encryption -> Compression -> mysqldump
	encryptionEnabled := s.BackupOptions.Encryption.Enabled
	if encryptionEnabled {
		encryptionKey := s.BackupOptions.Encryption.Key

		// Resolve encryption key jika belum ada
		if encryptionKey == "" {
			resolvedKey, source, err := encrypt.ResolveEncryptionKey("")
			if err != nil {
				return fmt.Errorf("gagal mendapatkan kunci enkripsi: %w", err)
			}
			encryptionKey = resolvedKey
			s.Logger.Infof("Kunci enkripsi diperoleh dari: %s", source)
		}

		// Buat encrypting writer dengan format kompatibel OpenSSL
		encryptingWriter, err := encrypt.NewEncryptingWriter(writer, []byte(encryptionKey))
		if err != nil {
			return fmt.Errorf("gagal membuat encrypting writer: %w", err)
		}
		closers = append(closers, encryptingWriter)
		writer = encryptingWriter

		s.Logger.Info("Enkripsi AES-256-GCM diaktifkan untuk backup (kompatibel dengan OpenSSL)")
	}

	// Setup kompresi jika diperlukan (layer kedua: di atas enkripsi)
	if compressionRequired {
		compressionConfig := compress.CompressionConfig{
			Type:  compress.CompressionType(compressionType),
			Level: compress.CompressionLevel(s.BackupAll.BackupOptions.Compression.Level),
		}

		compressingWriter, err := compress.NewCompressingWriter(writer, compressionConfig)
		if err != nil {
			return fmt.Errorf("gagal membuat compressing writer: %w", err)
		}
		closers = append(closers, compressingWriter)
		writer = compressingWriter

		s.Logger.Infof("Kompresi %s diaktifkan (level: %s)", compressionType, s.BackupAll.BackupOptions.Compression.Level)
	}

	// Defer semua closers dalam urutan terbalik
	defer func() {
		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i].Close(); err != nil {
				s.Logger.Errorf("Error closing writer: %v", err)
			}
		}
	}()

	// Buat command mysqldump
	cmd := exec.CommandContext(ctx, "mysqldump", mysqldumpArgs...)

	// Set stdout ke writer (file atau compressed writer)
	cmd.Stdout = writer
	cmd.Stderr = os.Stderr

	s.Logger.Info("Menjalankan mysqldump...")
	// Log command dengan password disembunyikan untuk keamanan
	logArgs := s.sanitizeArgsForLogging(mysqldumpArgs)
	s.Logger.Debugf("Command: mysqldump %s", strings.Join(logArgs, " "))

	// Jalankan command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysqldump gagal: %w", err)
	}

	s.Logger.Info("Mysqldump berhasil dijalankan")
	return nil
}
