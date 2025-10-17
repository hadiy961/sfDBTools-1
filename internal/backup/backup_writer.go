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

// executeMysqldumpWithPipe menjalankan mysqldump dengan pipe untuk kompresi dan enkripsi.
// Mengembalikan error untuk fatal errors dan stderr output untuk warnings/non-fatal errors
func (s *Service) executeMysqldumpWithPipe(ctx context.Context, mysqldumpArgs []string, outputPath string, compressionRequired bool, compressionType string) (string, error) {
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return "", fmt.Errorf("gagal membuat file output: %w", err)
	}
	defer outputFile.Close()

	var writer io.Writer = outputFile
	var closers []io.Closer

	// Urutan layer: mysqldump -> Compression -> Encryption -> File
	if s.BackupOptions.Encryption.Enabled {
		encryptionKey := s.BackupOptions.Encryption.Key
		if encryptionKey == "" {
			resolvedKey, source, err := encrypt.ResolveEncryptionKey("")
			if err != nil {
				return "", fmt.Errorf("gagal mendapatkan kunci enkripsi: %w", err)
			}
			encryptionKey = resolvedKey
			s.Logger.Infof("Kunci enkripsi diperoleh dari: %s", source)
		}

		encryptingWriter, err := encrypt.NewEncryptingWriter(writer, []byte(encryptionKey))
		if err != nil {
			return "", fmt.Errorf("gagal membuat encrypting writer: %w", err)
		}
		closers = append(closers, encryptingWriter)
		writer = encryptingWriter
	}

	if compressionRequired {
		compressionConfig := compress.CompressionConfig{
			Type:  compress.CompressionType(compressionType),
			Level: compress.CompressionLevel(s.BackupOptions.Compression.Level),
		}
		compressingWriter, err := compress.NewCompressingWriter(writer, compressionConfig)
		if err != nil {
			return "", fmt.Errorf("gagal membuat compressing writer: %w", err)
		}
		closers = append(closers, compressingWriter)
		writer = compressingWriter
	}

	defer func() {
		for i := len(closers) - 1; i >= 0; i-- {
			if err := closers[i].Close(); err != nil {
				s.Logger.Errorf("Error closing writer: %v", err)
			}
		}
	}()

	cmd := exec.CommandContext(ctx, "mysqldump", mysqldumpArgs...)
	cmd.Stdout = writer

	// Capture stderr untuk menangkap warnings dan errors
	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	// logArgs := s.sanitizeArgsForLogging(mysqldumpArgs)
	// s.Logger.Infof("Command: mysqldump %s", strings.Join(logArgs, " "))

	if err := cmd.Run(); err != nil {
		stderrOutput := stderrBuf.String()
		// Cek apakah ini error fatal atau hanya warning
		if s.isFatalMysqldumpError(err, stderrOutput) {
			return stderrOutput, fmt.Errorf("mysqldump gagal: %w", err)
		}
		// Jika bukan fatal error, kembalikan stderr sebagai warning
		return stderrOutput, nil
	}

	return stderrBuf.String(), nil
}
