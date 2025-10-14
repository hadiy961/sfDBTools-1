package backup

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sfDBTools/pkg/compress"
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

	// Setup kompresi jika diperlukan
	if compressionRequired {
		compressionConfig := compress.CompressionConfig{
			Type:  compress.CompressionType(compressionType),
			Level: compress.CompressionLevel(s.BackupAll.BackupOptions.Compression.Level),
		}

		compressingWriter, err := compress.NewCompressingWriter(outputFile, compressionConfig)
		if err != nil {
			return fmt.Errorf("gagal membuat compressing writer: %w", err)
		}
		defer compressingWriter.Close()
		writer = compressingWriter
	}

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
