package fs

import (
	"os"
	"path/filepath"
)

// Cek apakah file atau direktori ada
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// CheckFileExists cek apakah file dengan nama tertentu ada di direktori yang diberikan
func CheckFileExists(dir, filename string) (bool, error) {
	// Gabungkan direktori dan nama file
	path := filepath.Join(dir, filename)
	// Cek keberadaan file
	_, err := os.Stat(path)
	if err == nil {
		return true, nil // File exists
	}
	if os.IsNotExist(err) {
		return false, nil // File does not exist
	}
	return false, err // Some other error
}

// CreateDirIfNotExist membuat direktori jika belum ada
func CreateDirIfNotExist(dir string) error {
	// Jika tidak ada, buat direktori
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Create directory if it does not exist
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

// CheckDirExists mengecek apakah direktori ada
func CheckDirExists(dir string) (bool, error) {
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}
