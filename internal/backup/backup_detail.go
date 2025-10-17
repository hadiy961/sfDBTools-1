package backup

import (
	"context"
	"sfDBTools/pkg/database"
)

func (s *Service) GetDBDetail(ctx context.Context, dbFiltered []string) error {
	if len(dbFiltered) > 0 {
		// Pastikan nama database unik
		uniqueDBs := make(map[string]bool)
		for _, dbName := range dbFiltered {
			uniqueDBs[dbName] = true
		}
		dbNames := make([]string, 0, len(uniqueDBs))
		for dbName := range uniqueDBs {
			dbNames = append(dbNames, dbName)
		}

		s.Logger.Info("Mengumpulkan detail informasi database")
		var targetClient *database.Client
		var err error
		targetClient, err = s.Client.ConnectToTargetDB(ctx)
		if err != nil {
			return err
		}

		s.DatabaseDetail, err = targetClient.GetDatabaseDetails(ctx, dbNames, s.DBConfigInfo.ServerDBConnection.Host, s.DBConfigInfo.ServerDBConnection.Port)
		if err != nil {
			s.Logger.Warnf("Gagal mengumpulkan detail database: %v", err)
		} else {
			s.Logger.Infof("Berhasil mengumpulkan detail untuk %d database dari tabel database_details.", len(s.DatabaseDetail))
		}
	} else {
		s.Logger.Warn("Tidak ada database untuk dikumpulkan detailnya sebelum backup.")
	}

	return nil
}
