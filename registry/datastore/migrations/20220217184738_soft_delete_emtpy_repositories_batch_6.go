package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	m := &Migration{
		Migration: &migrate.Migration{
			Id: "20220217184738_soft_delete_emtpy_repositories_batch_6",
			Up: []string{
				softDeleteEmptyRepositoriesBatchQuery(160001, 170000),
			},
			Down: []string{
				undoSoftDeleteEmptyRepositoriesBatchQuery(160001, 170000),
			},
		},
		PostDeployment: false,
	}

	allMigrations = append(allMigrations, m)
}
