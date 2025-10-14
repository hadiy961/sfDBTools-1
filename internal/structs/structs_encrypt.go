// File : internal/structs/structs_encrypt.go
// Deskripsi : Structs untuk command encrypt dan decrypt
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-14
// Last Modified : 2024-10-14
package structs

// EncryptFlags - Flags untuk command encrypt
type EncryptFlags struct {
	InputFile     string `flag:"input" env:"SFDB_ENCRYPT_INPUT" default:""`
	OutputFile    string `flag:"output" env:"SFDB_ENCRYPT_OUTPUT" default:""`
	EncryptionKey string `flag:"key" env:"SFDB_ENCRYPTION_KEY" default:""`
	Overwrite     bool   `flag:"overwrite" env:"SFDB_ENCRYPT_OVERWRITE" default:"false"`
}

// DecryptFlags - Flags untuk command decrypt
type DecryptFlags struct {
	InputFile     string `flag:"input" env:"SFDB_DECRYPT_INPUT" default:""`
	OutputFile    string `flag:"output" env:"SFDB_DECRYPT_OUTPUT" default:""`
	EncryptionKey string `flag:"key" env:"SFDB_ENCRYPTION_KEY" default:""`
	Overwrite     bool   `flag:"overwrite" env:"SFDB_DECRYPT_OVERWRITE" default:"false"`
}
