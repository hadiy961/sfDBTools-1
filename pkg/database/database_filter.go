// File : pkg/database/database_filter.go
// Deskripsi : General database filtering system untuk backup dan scanning
// Author : Hadiyatna Muflihun
// Tanggal : 16 Oktober 2025
// Last Modified : 16 Oktober 2025

package database

import (
	"context"
	"fmt"
	"sfDBTools/pkg/fs"
	"strings"
)

// SystemDatabases adalah canonical list dari database sistem MySQL/MariaDB
// Menggunakan map untuk O(1) lookup performance
var SystemDatabases = map[string]struct{}{
	"information_schema": {},
	"mysql":              {},
	"performance_schema": {},
	"sys":                {},
	"innodb":             {},
}

// FilterOptions berisi opsi untuk filtering database
type FilterOptions struct {
	ExcludeSystem    bool     // Exclude system databases (information_schema, mysql, etc)
	ExcludeDatabases []string // Blacklist - database yang harus di-exclude
	IncludeDatabases []string // Whitelist - hanya database ini yang diizinkan (priority tertinggi)
	IncludeFile      string   // Path ke file berisi whitelist database (satu per baris)
}

// FilterStats berisi statistik hasil filtering
type FilterStats struct {
	TotalFound     int // Total database yang ditemukan
	TotalIncluded  int // Total database yang included (hasil akhir)
	TotalExcluded  int // Total database yang excluded
	ExcludedSystem int // Excluded karena system database
	ExcludedByList int // Excluded karena ada di blacklist
	ExcludedByFile int // Excluded karena tidak ada di whitelist file
	ExcludedEmpty  int // Excluded karena nama kosong
}

// FilterDatabases mengambil dan memfilter daftar database dari server berdasarkan FilterOptions
// Returns: filtered database list, statistics, error
func FilterDatabases(ctx context.Context, client *Client, options FilterOptions) ([]string, *FilterStats, error) {
	// 1. Get database list from server
	allDatabases, err := client.GetDatabaseList(ctx, client)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal mengambil daftar database: %w", err)
	}

	stats := &FilterStats{
		TotalFound: len(allDatabases),
	}

	// 2. Load whitelist from file if specified (priority tertinggi)
	var whitelistFromFile []string
	if options.IncludeFile != "" {
		whitelistFromFile, err = fs.ReadLinesFromFile(options.IncludeFile)
		if err != nil {
			return nil, nil, fmt.Errorf("gagal membaca file whitelist %s: %w", options.IncludeFile, err)
		}
		// Clean whitelist
		whitelistFromFile = cleanDatabaseList(whitelistFromFile)
	}

	// 3. Merge whitelist: file takes priority, then IncludeDatabases
	var whitelist []string
	if len(whitelistFromFile) > 0 {
		whitelist = whitelistFromFile
	} else if len(options.IncludeDatabases) > 0 {
		whitelist = cleanDatabaseList(options.IncludeDatabases)
	}

	// 4. Clean blacklist
	blacklist := cleanDatabaseList(options.ExcludeDatabases)

	// 5. Filter databases
	filtered := make([]string, 0, len(allDatabases))
	for _, dbName := range allDatabases {
		dbName = strings.TrimSpace(dbName)

		// Check exclusion
		if shouldExcludeDatabase(dbName, whitelist, blacklist, options.ExcludeSystem, stats) {
			stats.TotalExcluded++
			continue
		}

		filtered = append(filtered, dbName)
	}

	stats.TotalIncluded = len(filtered)

	// Validate result
	if len(filtered) == 0 {
		return nil, stats, fmt.Errorf("tidak ada database yang valid setelah filtering (found: %d, excluded: %d)",
			stats.TotalFound, stats.TotalExcluded)
	}

	return filtered, stats, nil
}

// shouldExcludeDatabase menentukan apakah database harus di-exclude
// Returns true jika database harus di-exclude, false jika harus di-include
func shouldExcludeDatabase(dbName string, whitelist, blacklist []string, excludeSystem bool, stats *FilterStats) bool {
	// 1. Exclude empty names
	if dbName == "" {
		stats.ExcludedEmpty++
		return true
	}

	// 2. Whitelist has highest priority - if specified, only include databases in whitelist
	if len(whitelist) > 0 {
		if !containsDatabase(whitelist, dbName) {
			stats.ExcludedByFile++
			return true
		}
		return false // Database is in whitelist, include it (skip other checks)
	}

	// 3. Check blacklist
	if containsDatabase(blacklist, dbName) {
		stats.ExcludedByList++
		return true
	}

	// 4. Check system databases
	if excludeSystem && isSystemDatabase(dbName) {
		stats.ExcludedSystem++
		return true
	}

	// Database passed all filters
	return false
}

// isSystemDatabase memeriksa apakah database adalah system database
func isSystemDatabase(dbName string) bool {
	_, exists := SystemDatabases[strings.ToLower(dbName)]
	return exists
}

// containsDatabase memeriksa apakah database ada dalam list
func containsDatabase(list []string, dbName string) bool {
	for _, item := range list {
		if strings.EqualFold(item, dbName) { // Case-insensitive comparison
			return true
		}
	}
	return false
}

// cleanDatabaseList membersihkan list database dari whitespace dan entry kosong
func cleanDatabaseList(list []string) []string {
	cleaned := make([]string, 0, len(list))
	for _, dbName := range list {
		dbName = strings.TrimSpace(dbName)
		if dbName != "" {
			cleaned = append(cleaned, dbName)
		}
	}
	return cleaned
}
