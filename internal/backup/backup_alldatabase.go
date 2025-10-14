// File : internal/backup/backup_alldatabase.go
// Deskripsi : Implementasi logika untuk backup semua database
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-08
// Last Modified : 2024-10-08

package backup

import (
	"context"
	"log"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/ui"
)

// BackupAllDatabases melakukan backup semua database yang terdaftar
func (s *Service) BackupAllDatabases() error {
	ui.Headers("Backup Semua Database")

	// Check flag configuration file
	if err := s.CheckAndSelectConfigFile(); err != nil {
		return err
	}

	s.DisplayBackupAllOptions()

	ctx := context.Background()

	// Membuat klien baru dengan semua konfigurasi di atas
	client, err := database.InitializeDatabase(s.DBConfigInfo.ServerDBConnection)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	//Pastikan koneksi ditutup di akhir
	defer client.Close()

	// Cek versi database
	version, err := client.GetVersion(ctx)
	if err != nil {
		s.Logger.Warn("Gagal mendapatkan versi database: " + err.Error())
		return err
	} else {
		s.Logger.Info("Terkoneksi ke database versi: " + version)
	}

	// Get, Set dan cek max_statement_time untuk sesi ini
	originalMaxStatementsTime, err := s.AturMaxStatementsTime(ctx, client)
	if err != nil {
		s.Logger.Warn("Gagal mengatur max_statement_time: " + err.Error())
		return err
	}

	// Check flag capture GTID
	if err := s.CaptureGTIDIfNeeded(ctx, client); err != nil {
		s.Logger.Warn("Gagal menangani opsi capture GTID: " + err.Error())
		return err
	}

	// Cek dan filter database yang akan di-backup
	dbFiltered, err := s.GetAndFilterDatabases(ctx, client)
	if err != nil {
		s.Logger.Error("Gagal mendapatkan dan memfilter database: " + err.Error())
		// Kembalikan nilai awal max_statement_time jika ada error
		s.KembalikanMaxStatementsTime(ctx, client, originalMaxStatementsTime)
		return err
	}

	// Database filter sudah siap untuk di-backup
	s.Logger.Info("Database filtering selesai, siap memulai proses backup") // Lakukan dump semua database yang sudah difilter
	if err := s.DumpAllDB(ctx, dbFiltered, "all"); err != nil {
		s.Logger.Error("Proses backup semua database gagal: " + err.Error())
		// Kembalikan nilai awal max_statement_time jika ada error
		s.KembalikanMaxStatementsTime(ctx, client, originalMaxStatementsTime)
		return err
	}

	// Kembalikan nilai awal
	s.KembalikanMaxStatementsTime(ctx, client, originalMaxStatementsTime)

	ui.PrintSuccess("Proses backup semua database selesai.")

	return nil
}
