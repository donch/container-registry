package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	m := &Migration{
		Migration: &migrate.Migration{
			Id: "20220216163253_soft_delete_emtpy_repositories_batch_3",
			Up: []string{
				softDeleteEmptyRepositoriesBatchQuery(50001, 60000),
			},
			Down: []string{
				undoSoftDeleteEmptyRepositoriesBatchQuery(50001, 60000),
			},
		},
		PostDeployment: false,
	}

	allMigrations = append(allMigrations, m)
}
