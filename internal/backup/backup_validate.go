package backup

import (
	"sfDBTools/pkg/fs"
	"sfDBTools/pkg/ui"
)

// ValidateOutput membuat direktori output jika belum ada
func (s *Service) ValidateOutput() error {
	// Buat direktori output jika belum ada
	ui.PrintSubHeader("Validasi Output Backup")

	// Membuat direktori output jika belum ada
	OutputDir, err := fs.CreateOutputDirs(s.BackupAll.BackupOptions.OutputDirectory, s.Config.Backup.Output.Structure.CreateSubdirs, s.Config.Backup.Output.Structure.Pattern, s.Config.General.ClientCode)
	if err != nil {
		return err
	}

	// Generate file path pattern untuk backup
	OutputFilePattern, err := fs.GenerateBackupFilename(s.Config.Backup.Output.Naming.Pattern, "all_databases", "full", s.Config.Backup.Output.Naming.IncludeClientCode, s.Config.Backup.Output.Naming.IncludeHostname, s.Config.General.ClientCode, "")
	if err != nil {
		return err
	}

	// Check disk space if enabled

	s.BackupAll.BackupOptions.OutputDirectory = OutputDir
	s.BackupAll.BackupOptions.OutputFile = OutputFilePattern

	// Tampilkan path lengkap file output
	s.Logger.Info("Direktori output siap: " + OutputDir)
	s.Logger.Info("File backup akan disimpan dengan nama: " + OutputFilePattern)
	return nil
}
