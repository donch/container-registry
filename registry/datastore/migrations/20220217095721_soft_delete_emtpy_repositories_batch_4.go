package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	m := &Migration{
		Migration: &migrate.Migration{
			Id: "20220217095721_soft_delete_emtpy_repositories_batch_4",
			Up: []string{
				softDeleteEmptyRepositoriesBatchQuery(90001, 100000),
			},
			Down: []string{
				undoSoftDeleteEmptyRepositoriesBatchQuery(90001, 100000),
			},
		},
		PostDeployment: false,
	}

	allMigrations = append(allMigrations, m)
}
