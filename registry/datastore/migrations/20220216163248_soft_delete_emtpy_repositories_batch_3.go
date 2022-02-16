package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	m := &Migration{
		Migration: &migrate.Migration{
			Id: "20220216163248_soft_delete_emtpy_repositories_batch_3",
			Up: []string{
				softDeleteEmptyRepositoriesBatchQuery(30001, 40000),
			},
			Down: []string{
				undoSoftDeleteEmptyRepositoriesBatchQuery(30001, 40000),
			},
		},
		PostDeployment: false,
	}

	allMigrations = append(allMigrations, m)
}
