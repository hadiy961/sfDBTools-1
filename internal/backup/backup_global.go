package backup

import (
	"context"
	"fmt"
	"sfDBTools/internal/structs"
	"sfDBTools/pkg/database"
	"sfDBTools/pkg/fs"
	"sfDBTools/pkg/ui"
	"strings"
)

// KembalikanMaxStatementsTime mengembalikan nilai max_statements_time saat ini dari sesi database.
func (s *Service) KembalikanMaxStatementsTime(ctx context.Context, client *database.Client, original float64) {
	ui.PrintSubHeader("Mengembalikan nilai max_statement_time")
	if err := client.SetMaxStatementsTime(ctx, original); err != nil {
		s.Logger.Warn("Gagal mengembalikan nilai max_statement_time: " + err.Error())
	} else {
		s.Logger.Info("Nilai max_statement_time berhasil dikembalikan ke " + fmt.Sprintf("%f", original))
	}
}

// AturMaxStatementsTimeJikaPerlu mengatur max_statements_time jika opsi diatur.
// Mengembalikan nilai asli untuk dikembalikan nanti.
func (s *Service) AturMaxStatementsTime(ctx context.Context, client *database.Client) (float64, error) {
	// Set dan cek max_statement_time untuk sesi ini
	ui.PrintSubHeader("Setting max_statement_time")
	// Cek nilai awal
	originalMaxStatementsTime, err := client.GetMaxStatementsTime(ctx)
	if err != nil {
		s.Logger.Warn("Gagal mendapatkan timeout statement awal: " + err.Error())
	} else {
		s.Logger.Info(fmt.Sprintf("Original max_statement_time: %f detik", originalMaxStatementsTime))
	}

	// Set nilai baru
	s.Logger.Info("Mengatur max_statement_time ke 0 detik untuk sesi ini...")
	// 0 berarti tidak ada batasan waktu
	if err := client.SetMaxStatementsTime(ctx, 0); err != nil {
		s.Logger.Warn("Gagal mengatur timeout statement: " + err.Error())
	}

	// Verifikasi nilai baru
	currentTimeout, err := client.GetMaxStatementsTime(ctx)
	if err != nil {
		s.Logger.Warn("Gagal mendapatkan timeout statement saat ini: " + err.Error())
	} else {
		s.Logger.Debug(fmt.Sprintf("Current max_statement_time: %f detik", currentTimeout))
	}

	return originalMaxStatementsTime, nil
}

// buildMysqldumpArgs membangun argumen mysqldump dengan kredensial database
// Parameter singleDB: jika tidak kosong, akan backup database tunggal tersebut
// Parameter dbFiltered: list database untuk backup multiple (diabaikan jika singleDB diisi)
func (s *Service) buildMysqldumpArgs(baseDumpArgs string, dbFiltered []string, singleDB string) []string {
	var args []string

	// Tambahkan kredensial database
	dbConn := s.BackupOptions.DBConfig.ServerDBConnection

	// Host
	if dbConn.Host != "" {
		args = append(args, "--host="+dbConn.Host)
	}

	// Port
	if dbConn.Port != 0 {
		args = append(args, fmt.Sprintf("--port=%d", dbConn.Port))
	}

	// User
	if dbConn.User != "" {
		args = append(args, "--user="+dbConn.User)
	}

	// Password
	if dbConn.Password != "" {
		args = append(args, "--password="+dbConn.Password)
	}

	// Tambahkan argumen mysqldump dari konfigurasi
	if baseDumpArgs != "" {
		baseArgs := strings.Fields(baseDumpArgs)
		args = append(args, baseArgs...)
	}

	// Mode single database
	if singleDB != "" {
		args = append(args, "--databases")
		args = append(args, singleDB)
		return args
	}

	// Mode multiple databases
	totalDB := s.FilterInfo.TotalDatabases
	excludedDB := s.FilterInfo.ExcludedDatabases
	toBackupDB := s.FilterInfo.IncludedDatabases

	// Tentukan apakah menggunakan --all-databases atau --databases
	if totalDB == toBackupDB && excludedDB == 0 {
		args = append(args, "--all-databases")
		s.Logger.Infof("Menggunakan --all-databases untuk backup semua database (%d database)", totalDB)
	} else {
		args = append(args, "--databases")
		args = append(args, dbFiltered...)
		s.Logger.Infof("Menggunakan --databases dengan %d database yang dipilih", toBackupDB)
	}

	return args
}

// GetAndFilterDatabases mendapatkan dan memfilter database berdasarkan opsi exclude
func (s *Service) GetAndFilterDatabases(ctx context.Context, client *database.Client) ([]string, error) {
	ui.PrintSubHeader("Mendapatkan dan Memfilter Database")

	// 1. Dapatkan daftar database dari server
	s.Logger.Info("Mendapatkan daftar database dari server...")
	allDatabases, err := client.GetDatabaseList(ctx, client)
	if err != nil {
		s.Logger.Error("Gagal mendapatkan daftar database: " + err.Error())
		return nil, err
	}

	totalFound := len(allDatabases)
	s.Logger.Infof("Ditemukan %d database di server.", totalFound)

	// 2. Filter database
	var validDatabases []string
	var excludedCount int

	for _, dbName := range allDatabases {
		dbName = strings.TrimSpace(dbName)
		if dbName == "" {
			excludedCount++
			continue
		}

		if s.BackupFilterDatabase(dbName) {
			excludedCount++
			continue
		}

		validDatabases = append(validDatabases, dbName)
	}

	// 3. Tampilkan statistik sederhana
	s.Logger.Infof("Total database: %d, Untuk backup: %d, Dikecualikan: %d",
		totalFound, len(validDatabases), excludedCount)

	// 4. Validasi hasil
	if len(validDatabases) == 0 {
		s.Logger.Error("Tidak ada database yang valid untuk di-backup setelah filtering")
		return nil, fmt.Errorf("tidak ada database untuk di-backup dari %d database yang ditemukan", totalFound)
	}

	if excludedCount == 0 {
		s.Logger.Info("Tidak ada database yang dikecualikan; semua database akan di-backup.")
	}
	// Simpan statistik filtering ke struct
	s.FilterInfo = &structs.FilterInfo{
		TotalDatabases:    totalFound,
		ExcludedDatabases: excludedCount,
		IncludedDatabases: len(validDatabases),
		SystemDatabases:   totalFound - len(validDatabases) - excludedCount,
	}

	return validDatabases, nil
}

// BackupFilterDatabase memeriksa apakah sebuah database harus dikecualikan berdasarkan opsi exclude.
// Mengembalikan true jika database harus dikecualikan, false jika harus disertakan.
func (s *Service) BackupFilterDatabase(dbName string) bool {
	// Validasi input
	if strings.TrimSpace(dbName) == "" {
		return true // Exclude database dengan nama kosong
	}

	// 1. Jika ada file whitelist, hanya database dalam file yang diizinkan
	if s.BackupOptions.DBList != "" {
		allowedDBs, err := fs.ReadLinesFromFile(s.BackupOptions.DBList)
		if err != nil {
			s.Logger.Warnf("Gagal membaca file db_list: %v", err)
			return true // Exclude semua jika file gagal dibaca
		}

		// Cek apakah database ada dalam whitelist
		for _, allowedDB := range allowedDBs {
			if strings.TrimSpace(allowedDB) == dbName {
				return false // Database diizinkan
			}
		}
		return true // Database tidak ada dalam whitelist, exclude
	}

	// 2. Cek blacklist database yang dikecualikan
	for _, excludeDB := range s.BackupOptions.Exclude.Databases {
		if strings.TrimSpace(excludeDB) == dbName {
			return true // Exclude database dalam blacklist
		}
	}

	// 3. Cek sistem database jika opsi exclude system aktif
	if s.BackupOptions.Exclude.SystemsDB {
		systemDBs := map[string]bool{
			"information_schema": true,
			"mysql":              true,
			"performance_schema": true,
			"sys":                true,
			"innodb":             true,
		}
		if systemDBs[dbName] {
			return true // Exclude sistem database
		}
	}

	// Database tidak dikecualikan
	return false
}
