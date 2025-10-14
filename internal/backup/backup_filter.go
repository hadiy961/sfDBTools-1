// File: internal/backup/backup_filter.go
// Deskripsi: Logika untuk memfilter database berdasarkan opsi exclude pada modul backup
// Author: Hadiyatna Muflihun
// Tanggal: 2024-10-08
// Last Modified: 2024-10-08

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
