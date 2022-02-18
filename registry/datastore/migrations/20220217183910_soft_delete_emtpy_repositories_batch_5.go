package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	m := &Migration{
		Migration: &migrate.Migration{
			Id: "20220217183910_soft_delete_emtpy_repositories_batch_5",
			Up: []string{
				softDeleteEmptyRepositoriesBatchQuery(130001, 140000),
			},
			Down: []string{
				undoSoftDeleteEmptyRepositoriesBatchQuery(130001, 140000),
			},
		},
		PostDeployment: false,
	}

	allMigrations = append(allMigrations, m)
}
