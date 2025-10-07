// File : internal/default_value/default_dbconfig.go
// Deskripsi : Nilai default untuk flags pada modul dbconfig
// Author : Hadiyatna Muflihun
// Tanggal : 2024-10-03
// Last Modified : 2024-10-03
package defaultvalue

import (
	"sfDBTools/internal/structs"
)

// GetDefaultDBConfigCreate returns default values for DBConfigCreateFlags
func GetDefaultDBConfigCreate() (*structs.DBConfigCreateFlags, error) {
	return &structs.DBConfigCreateFlags{
		DBConfigInfo: structs.DBConfigInfo{
			ServerDBConnection: structs.ServerDBConnection{
				Host:     "localhost",
				Port:     3306,
				User:     "root",
				Password: "",
			},
			EncryptionKey: "mysecretkey",
			ConfigName:    "local_mariadb",
		},
		Interactive: true,
	}, nil
}

// GetDefaultDBConfigEdit returns default values for DBConfigEditFlags
func GetDefaultDBConfigEdit() (*structs.DBConfigEditFlags, error) {
	return &structs.DBConfigEditFlags{
		File: "local_mariadb",
		DBConfigInfo: structs.DBConfigInfo{
			ServerDBConnection: structs.ServerDBConnection{
				Host:     "localhost",
				Port:     3306,
				User:     "root",
				Password: "",
			},
			EncryptionKey: "",
			ConfigName:    "local_mariadb",
		},
		Interactive: true,
	}, nil
}
