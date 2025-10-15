// File : internal/dbscan/dbscan_main.go
// Deskripsi : Service utama untuk database scanning
// Author : Hadiyatna Muflihun
// Tanggal : 15 Oktober 2025
// Last Modified : 15 Oktober 2025

package dbscan

import (
	"sfDBTools/internal/appconfig"
	"sfDBTools/internal/structs"

	"github.com/sirupsen/logrus"
)

// Service adalah service untuk database scanning
type Service struct {
	Logger       *logrus.Logger
	Config       *appconfig.Config
	ScanOptions  structs.ScanOptions
	DBConfigInfo structs.DBConfigInfo
}

// NewService membuat instance baru dari Service
func NewService(logger *logrus.Logger, config *appconfig.Config) *Service {
	return &Service{
		Logger: logger,
		Config: config,
	}
}

// SetScanOptions mengatur opsi scan
func (s *Service) SetScanOptions(opts structs.ScanOptions) {
	s.ScanOptions = opts
	s.DBConfigInfo = opts.DBConfig
}
