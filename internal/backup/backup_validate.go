package backup

import (
	"sfDBTools/pkg/fs"
)

// ValidateOutput membuat direktori output jika belum ada
func (s *Service) ValidateOutput() error {

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

	// Update struct dengan path yang sudah divalidasi
	s.BackupOptions.OutputDirectory = OutputDir
	s.BackupOptions.OutputFile = OutputFilePattern
	return nil
}
