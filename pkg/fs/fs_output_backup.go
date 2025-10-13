package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// BuildSubdirPath membangun path subdirektori sesuai pattern.
// Pattern bersifat dinamis dan mendukung token berikut:
// - {date} atau {date:FORMAT} -> default FORMAT = "2006-01-02"
// - {timestamp} atau {timestamp:FORMAT} -> default FORMAT = "20060102T150405"
// - {year}, {month}, {day}
// - field lain yang disediakan lewat vars map, mis. {client}, {client_code}, {hostname}
// Contoh pattern: "{date}/{client}" atau "{date:2006/01}/{client_code}"
func BuildSubdirPath(structurePattern, client string) (string, error) {
	if structurePattern == "" {
		return "", fmt.Errorf("empty structure pattern")
	}
	vars := map[string]string{
		"client": client,
	}
	return BuildSubdirPathFromPattern(structurePattern, vars, time.Now())
}

// BuildSubdirPathFromPattern menggantikan token-token dalam pola menggunakan vars dan waktu sekarang.
// Format token: {nama} atau {nama:format}. Untuk token 'date' dan 'timestamp' format mengikuti Go time.Format.
func BuildSubdirPathFromPattern(pattern string, vars map[string]string, now time.Time) (string, error) {
	if pattern == "" {
		return "", fmt.Errorf("pola struktur kosong")
	}

	// Validasi dasar sebelum penggantian token
	if err := ValidateSubdirPattern(pattern, vars); err != nil {
		return "", fmt.Errorf("pola struktur tidak valid: %w", err)
	}

	// regex menangkap {name} atau {name:format}
	re := regexp.MustCompile(`\{([a-zA-Z0-9_]+)(?::([^}]+))?\}`)

	replaced := re.ReplaceAllStringFunc(pattern, func(tok string) string {
		parts := re.FindStringSubmatch(tok)
		if len(parts) < 2 {
			return ""
		}
		name := parts[1]
		format := ""
		if len(parts) >= 3 {
			format = parts[2]
		}

		switch name {
		case "date":
			if format == "" {
				format = "2006-01-02"
			}
			return now.Format(format)
		case "timestamp":
			if format == "" {
				// default format: YYYYMMDD_HHMMSS
				format = "20060102_150405"
			}
			return now.Format(format)
		case "year":
			return fmt.Sprintf("%04d", now.Year())
		case "month":
			return fmt.Sprintf("%02d", int(now.Month()))
		case "day":
			return fmt.Sprintf("%02d", now.Day())
		default:
			if v, ok := vars[name]; ok {
				return v
			}
			// jika tidak ada, kosongkan token
			return ""
		}
	})

	// sanitasi path dan cegah path-traversal
	cleaned := filepath.Clean(replaced)
	// jika hasil menjadi '..' atau memiliki ../ di awal, buang bagian tersebut
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
		cleaned = strings.ReplaceAll(cleaned, "..", "")
	}
	// pastikan tidak ada absolute path (hindari memaksa root)
	if filepath.IsAbs(cleaned) {
		cleaned = strings.TrimPrefix(cleaned, string(os.PathSeparator))
	}

	return cleaned, nil
}

// ValidateSubdirPattern memastikan pola valid:
// - jumlah '{' sama dengan jumlah '}'
// - token hanya mengandung nama yang diizinkan (builtin atau vars)
// - tidak memulai dengan '/' (absolute path)
// - tidak mengandung path traversal '..'
func ValidateSubdirPattern(pattern string, vars map[string]string) error {
	if strings.Count(pattern, "{") != strings.Count(pattern, "}") {
		return fmt.Errorf("kurung kurawal tidak seimbang pada pola")
	}
	if pattern == "" {
		return fmt.Errorf("pola kosong")
	}
	if strings.HasPrefix(pattern, string(os.PathSeparator)) {
		return fmt.Errorf("pola tidak boleh absolut (tidak boleh diawali dengan %q)", string(os.PathSeparator))
	}
	if strings.Contains(pattern, "..") {
		return fmt.Errorf("pola tidak boleh mengandung path traversal '..'")
	}

	// allowed builtin tokens
	builtins := map[string]bool{
		"date":      true,
		"timestamp": true,
		"year":      true,
		"month":     true,
		"day":       true,
	}

	// regex menangkap {name} atau {name:format}
	re := regexp.MustCompile(`\{([a-zA-Z0-9_]+)(?::([^}]+))?\}`)
	matches := re.FindAllStringSubmatch(pattern, -1)
	for _, m := range matches {
		if len(m) < 2 {
			return fmt.Errorf("format token tidak valid")
		}
		name := m[1]
		if builtins[name] {
			continue
		}
		// jika bukan builtin, pastikan ada di vars
		if _, ok := vars[name]; !ok {
			return fmt.Errorf("token tidak dikenal '%s' di pola; token yang diizinkan: date,timestamp,year,month,day atau kunci yang tersedia di vars", name)
		}
	}
	return nil
}

// CreateOutputDirs membuat direktori base dan (opsional) subdirektori berdasarkan konfigurasi.
// Mengembalikan path final tempat backup akan disimpan.
func CreateOutputDirs(baseDir string, createSubdirs bool, structurePattern, client string) (string, error) {
	if baseDir == "" {
		return "", fmt.Errorf("direktori dasar kosong")
	}

	// Pastikan direktori dasar ada
	dir, err := CheckDirExists(baseDir)
	if err != nil {
		return "", fmt.Errorf("gagal memastikan direktori dasar: %w", err)
	}
	if !dir {
		fmt.Println("Membuat direktori dasar:", baseDir)
		if err := CreateDirIfNotExist(baseDir); err != nil {
			return "", fmt.Errorf("gagal membuat direktori dasar: %w", err)
		}
	}

	finalDir := baseDir
	if createSubdirs {
		subdir, err := BuildSubdirPath(structurePattern, client)
		if err != nil {
			return "", fmt.Errorf("gagal membangun path subdirektori: %w", err)
		}
		finalDir = filepath.Join(baseDir, subdir)
		if err := CreateDirIfNotExist(finalDir); err != nil {
			return "", fmt.Errorf("gagal membuat path subdirektori: %w", err)
		}
	}
	return finalDir, nil
}

// GenerateBackupFilename membentuk nama file backup sesuai pola penamaan.
// Pola mendukung token: {database}, {timestamp}, {type}, {client_code}, {hostname}
// Default {timestamp} diformat sebagai YYYYMMDD_HHMMSS (contoh: 20251013_141312).
func GenerateBackupFilename(namingPattern, database, typ string, includeClientCode bool, includeHostname bool, clientCode, hostname string) (string, error) {
	if namingPattern == "" {
		return "", fmt.Errorf("pola penamaan kosong")
	}
	if database == "" {
		return "", fmt.Errorf("nama database diperlukan untuk membuat nama file")
	}
	now := time.Now()
	// default format timestamp: YYYYMMDD_HHMMSS
	timestamp := now.Format("20060102_150405")

	p := namingPattern
	p = strings.ReplaceAll(p, "{database}", database)
	p = strings.ReplaceAll(p, "{timestamp}", timestamp)
	p = strings.ReplaceAll(p, "{type}", typ)

	// optional tokens
	if includeClientCode {
		p = strings.ReplaceAll(p, "{client_code}", clientCode)
	} else {
		p = strings.ReplaceAll(p, "{client_code}", "")
	}
	if includeHostname {
		p = strings.ReplaceAll(p, "{hostname}", hostname)
	} else {
		p = strings.ReplaceAll(p, "{hostname}", "")
	}

	// collapse multiple underscores or separators that might occur due to empty tokens
	p = sanitizeFilename(p)
	return p, nil
}

// sanitizeFilename melakukan pembersihan sederhana pada nama file:
// - mengganti spasi dengan underscore
// - menghapus double-underscore
// - membersihkan path traversal
func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, " ", "_")
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}
	name = strings.Trim(name, "_")
	name = filepath.Clean(name)
	// pastikan bukan path yang naik ke atas
	if name == ".." || strings.HasPrefix(name, ".."+string(os.PathSeparator)) {
		name = strings.ReplaceAll(name, "..", "")
	}
	// tidak mengizinkan path separator dalam nama file
	name = strings.ReplaceAll(name, string(os.PathSeparator), "_")
	return name
}

// CleanupTempDir menghapus semua isi dalam direktori temp (tetap mempertahankan direktori).
// Jika direktori tidak ada, tidak error.
func CleanupTempDir(tempDir string) error {
	if tempDir == "" {
		return fmt.Errorf("empty temp directory path")
	}
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		// jika tidak ada dir, anggap berhasil
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		p := filepath.Join(tempDir, e.Name())
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("failed remove %s: %w", p, err)
		}
	}
	return nil
}
