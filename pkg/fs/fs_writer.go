// File : pkg/fs/fs_writer.go
// Deskripsi : Fungsi utilitas untuk operasi filesystem
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package fs

import (
	"os"
)

// WriteFile menulis data ke file di path yang diberikan dengan permission 0600.
func WriteFile(filePath string, data []byte) error {
	return os.WriteFile(filePath, data, 0600)
}

// ReadDirFiles membaca nama-nama file dalam direktori yang diberikan.
func ReadDirFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() {
			files = append(files, e.Name())
		}
	}
	return files, nil
}

// ReadFile membaca isi file dari path yang diberikan.
func ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// RemoveFile menghapus file di path yang diberikan.
func RemoveFile(path string) error {
	return os.Remove(path)
}

// ReadLinesFromFile membaca semua baris dari file di path yang diberikan.
func ReadLinesFromFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := []string{}
	currentLine := ""
	for _, b := range data {
		if b == '\n' {
			lines = append(lines, currentLine)
			currentLine = ""
		} else {
			currentLine += string(b)
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines, nil
}
