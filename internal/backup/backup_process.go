package backup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sfDBTools/pkg/compress"
	"sfDBTools/pkg/ui"
)

// DumpAllDB adalah fungsi internal untuk melakukan backup semua database
func (s *Service) DumpAllDB(ctx context.Context, dbFiltered []string, mode string) error {

	ui.PrintSubHeader("Proses Backup")

	s.Logger.Info("Memulai proses backup database...")

	// Pastikan output directory sudah ada dengan memanggil ValidateOutput
	if err := s.ValidateOutput(); err != nil {
		return fmt.Errorf("gagal memvalidasi direktori output: %w", err)
	}

	// Mysqldump command arguments
	baseDumpArgs := s.Config.Backup.MysqlDumpArgs
	outputFile := s.BackupOptions.OutputFile + ".sql"
	outputDir := s.BackupOptions.OutputDirectory
	compressionType := s.BackupOptions.Compression.Type
	compressionRequired := s.BackupOptions.Compression.Enabled
	encryptionEnabled := s.BackupOptions.Encryption.Enabled

	if encryptionEnabled {
		s.Logger.Info("Enkripsi diaktifkan, memastikan ketersediaan kunci enkripsi...")
	} else {
		s.Logger.Info("Enkripsi tidak diaktifkan, melewati langkah kunci enkripsi...")
	}

	if compressionRequired {
		s.Logger.Infof("Kompresi diaktifkan dengan tipe: %s", compressionType)
	} else {
		s.Logger.Info("Kompresi tidak diaktifkan, melewati langkah kompresi...")
	}

	// Tambahkan ekstensi kompresi jika diperlukan
	if compressionRequired {
		extension := compress.GetFileExtension(compress.CompressionType(compressionType))
		outputFile += extension
	}

	// Pastikan direktori output ada
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori output: %w", err)
	}

	fullOutputPath := filepath.Join(outputDir, outputFile)

	// Siapkan argumen mysqldump dengan kredensial database
	mysqldumpArgs := s.buildMysqldumpArgs(baseDumpArgs, dbFiltered)

	s.Logger.Debug("Direktori output: " + outputDir)
	s.Logger.Debug("File output: " + fullOutputPath)

	// Jalankan mysqldump dengan pipe ke output file
	if err := s.executeMysqldumpWithPipe(ctx, mysqldumpArgs, fullOutputPath, compressionRequired, compressionType); err != nil {
		return fmt.Errorf("gagal menjalankan mysqldump: %w", err)
	}

	s.Logger.Info("Proses backup semua database selesai.")
	s.Logger.Infof("File backup tersimpan di: %s", fullOutputPath)

	return nil
}
