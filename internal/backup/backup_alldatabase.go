// File : internal/backup/backup_alldatabase.go
// Deskripsi : Implementasi logika untuk backup semua database
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-08
// Last Modified : 2024-10-08

package backup

import (
	"context"
	"errors"
	"log"
	"sfDBTools/pkg/common"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/encrypt"
	"sfDBTools/pkg/ui"
	"strconv"
)

// BackupAllDatabases melakukan backup semua database yang terdaftar
func (s *Service) BackupAllDatabases() error {

	ui.Headers("Backup Semua Database")
	ui.PrintSubHeader("Backup All Databases Options")
	// menampilkan seluruh opsi backup yang digunakan dalam bentuk table
	headers := []string{"Option", "Value"}
	port := s.BackupAll.BackupOptions.DBConfig.ServerDBConnection.Port
	data := [][]string{
		{"Configuration Path", s.BackupAll.BackupOptions.DBConfig.FilePath},
		{"Host", s.BackupAll.BackupOptions.DBConfig.ServerDBConnection.Host},
		{"Port", strconv.Itoa(port)},
		{"Username", s.BackupAll.BackupOptions.DBConfig.ServerDBConnection.User},
		{"Backup Directory", s.BackupAll.BackupOptions.OutputDirectory},
		{"Compression", s.BackupAll.BackupOptions.Compression.Type},
		{"Compression Level", (s.BackupAll.BackupOptions.Compression.Level)},
		{"Encryption Enabled", strconv.FormatBool(s.BackupAll.BackupOptions.Encryption.Enabled)},
		{"Cleanup Enabled", strconv.FormatBool(s.BackupAll.Cleanup.Enabled)},
		{"Cleanup Schedule", s.BackupAll.Cleanup.Scheduled},
		{"Retention Days", strconv.Itoa(s.BackupAll.Cleanup.RetentionDays)},
		{"Exclude Databases", ui.FormatStringSlice(s.BackupAll.Exclude.Databases)},
		{"Exclude Users", strconv.FormatBool(s.BackupAll.Exclude.Users)},
		{"Exclude System Databases", strconv.FormatBool(s.BackupAll.Exclude.SystemsDB)},
		{"Exclude Data", strconv.FormatBool(s.BackupAll.Exclude.Data)},
		{"Database List File", s.BackupAll.DBList.File},
		{"Verification Disk Check", strconv.FormatBool(s.BackupAll.Verification.DiskCheck)},
		{"Capture GTID", strconv.FormatBool(s.BackupAll.CaptureGtid)},
	}
	ui.FormatTable(headers, data)

	// Check flag configuration file
	if s.BackupAll.BackupOptions.DBConfig.FilePath == "" {
		// Jika tidak ada file konfigurasi, tampilkan pilihan interaktif
		ui.PrintWarning("Tidak ada file konfigurasi yang ditentukan. Menjalankan mode interaktif...")
		DBConfigInfo, err := encrypt.SelectExistingDBConfig("Pilih file konfigurasi database sumber:")
		if err != nil {
			if err == ErrUserCancelled {
				s.Logger.Warn("Proses backup dibatalkan oleh pengguna.")
				return ErrUserCancelled
			}
			s.Logger.Warn("Proses pemilihan file konfigurasi gagal: " + err.Error())
			return err
		}
		s.BackupAll.BackupOptions.DBConfig = DBConfigInfo
		s.DisplayConnectionInfo(DBConfigInfo)
	} else {
		abs, name, err := common.ResolveConfigPath(s.BackupAll.BackupOptions.DBConfig.FilePath)
		if err != nil {
			return err
		}
		var encryptionKey string
		if s.BackupAll.BackupOptions.Encryption.Key == "" {
			encryptionKey = s.BackupAll.BackupOptions.Encryption.Key
		} else {
			encryptionKey = s.BackupAll.BackupOptions.Encryption.Key
		}

		info, err := encrypt.LoadAndParseConfig(abs, encryptionKey)
		if err != nil {
			s.Logger.Warn("Gagal memuat isi detail konfigurasi untuk validasi: " + err.Error())
		}
		if info != nil {
			s.BackupAll.BackupOptions.DBConfig.ServerDBConnection.Host = info.ServerDBConnection.Host
			s.BackupAll.BackupOptions.DBConfig.ServerDBConnection.Port = info.ServerDBConnection.Port
			s.BackupAll.BackupOptions.DBConfig.ServerDBConnection.User = info.ServerDBConnection.User
			// Preserve password if flags didn't provide one
			if s.BackupAll.BackupOptions.DBConfig.ServerDBConnection.Password == "" {
				s.BackupAll.BackupOptions.DBConfig.ServerDBConnection.Password = info.ServerDBConnection.Password
			}
		}
		s.BackupAll.BackupOptions.DBConfig.FilePath = abs
		s.BackupAll.BackupOptions.DBConfig.ConfigName = name
		s.DisplayConnectionInfo(s.BackupAll.BackupOptions.DBConfig)
	}

	ctx := context.Background()

	// Membuat klien baru dengan semua konfigurasi di atas
	client, err := database.InitializeDatabase(s.BackupAll.BackupOptions.DBConfig.ServerDBConnection)
	if err != nil {
		// Menggunakan log.Fatalf akan menghentikan aplikasi jika koneksi gagal,
		// ini adalah pola yang umum untuk service utama.
		log.Fatalf("Error: %v", err)
	}
	//Pastikan koneksi ditutup di akhir
	defer client.Close()

	// Cek versi database
	version, err := client.GetVersion(ctx)
	if err != nil {
		s.Logger.Warn("Gagal mendapatkan versi database: " + err.Error())
	} else {
		s.Logger.Info("Terkoneksi ke database versi: " + version)
	}

	// Get, Set dan cek max_statement_time untuk sesi ini
	originalMaxStatementsTime, err := s.AturMaxStatementsTime(ctx, client)
	if err != nil {
		s.Logger.Warn("Gagal mengatur max_statement_time: " + err.Error())
	}

	// Check flag capture GTID
	if s.BackupAll.CaptureGtid {
		// Cek dukungan GTID
		enabled, pos, err := s.GetGTID(ctx, client)

		if err != nil {
			if errors.Is(err, ErrGTIDUnsupported) {
				s.Logger.Warn("Server tidak mendukung GTID: " + err.Error())
			} else if errors.Is(err, ErrGTIDPermissionDenied) {
				s.Logger.Warn("Tidak memiliki izin untuk membaca variabel GTID: " + err.Error())
			} else {
				s.Logger.Warn("Gagal memeriksa dukungan GTID: " + err.Error())
			}
			s.Logger.Warn("Menonaktifkan opsi capture GTID.")
			s.BackupAll.CaptureGtid = false
		}

		if enabled {
			s.Logger.Info("GTID saat ini: " + pos)
			// Simpan posisi GTID awal untuk referensi
			s.BackupAll.BackupInfo.GTIDCaptured = pos
		} else {
			s.Logger.Warn("GTID tidak diaktifkan pada server ini.")
			s.Logger.Warn("Menonaktifkan opsi capture GTID.")
			s.BackupAll.CaptureGtid = false
		}
	}

	// Panggil fungsi utama untuk backup semua database
	if err := s.backupAllDatabases(ctx, client); err != nil {
		s.Logger.Error("Proses backup semua database gagal: " + err.Error())
		// Kembalikan nilai awal max_statement_time jika ada error
		s.KembalikanMaxStatementsTime(ctx, client, originalMaxStatementsTime)
		return err
	}

	// Jika ada opsi cleanup, jalankan proses cleanup
	if s.BackupAll.Cleanup.Enabled {
		s.Logger.Info("Memulai proses cleanup backup lama...")
		if err := s.cleanupOldBackups(); err != nil {
			s.Logger.Warn("Proses cleanup backup lama gagal: " + err.Error())
		} else {
			s.Logger.Info("Proses cleanup backup lama selesai.")
		}
	}

	// Verifikasi file backup di disk jika diaktifkan
	if s.BackupAll.Verification.DiskCheck {
		ui.PrintSubHeader("Verifikasi File Backup")
		s.Logger.Info("Memulai verifikasi file backup di disk...")
		if err := s.verifyBackupFilesOnDisk(); err != nil {
			s.Logger.Warn("Verifikasi file backup di disk gagal: " + err.Error())
		} else {
			s.Logger.Info("Verifikasi file backup di disk selesai.")
		}
	}

	// Kembalikan nilai awal
	s.KembalikanMaxStatementsTime(ctx, client, originalMaxStatementsTime)

	ui.PrintSuccess("Proses backup semua database selesai.")

	return nil
}

// backupAllDatabases adalah fungsi internal untuk melakukan backup semua database
func (s *Service) backupAllDatabases(ctx context.Context, client *database.Client) error {
	ui.PrintSubHeader("Proses Backup")

	// Panggil fungsi backup semua database
	s.Logger.Info("Memulai proses backup semua database...")

	// 1. Dapatkan daftar database
	dbs, err := s.getDatabaseList(ctx, client)
	if err != nil {
		return err
	}
	if len(dbs) == 0 {
		return errors.New("tidak ada database yang ditemukan untuk dibackup")
	}
	s.Logger.Info("Ditemukan " + strconv.Itoa(len(dbs)) + " database untuk dibackup.")

	return nil
}

// cleanupOldBackups menghapus file backup lama berdasarkan kebijakan retensi.
func (s *Service) cleanupOldBackups() error {
	// Implementasi logika cleanup file backup lama
	s.Logger.Info("Fungsi cleanupOldBackups belum diimplementasikan.")
	return nil
}

// verifyBackupFilesOnDisk memverifikasi keberadaan dan integritas file backup di disk.
func (s *Service) verifyBackupFilesOnDisk() error {
	// Implementasi logika verifikasi file backup di disk
	s.Logger.Info("Fungsi verifyBackupFilesOnDisk belum diimplementasikan.")
	return nil
}
