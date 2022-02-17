package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	m := &Migration{
		Migration: &migrate.Migration{
			Id: "20220217183900_soft_delete_emtpy_repositories_batch_5",
			Up: []string{
				softDeleteEmptyRepositoriesBatchQuery(100001, 110000),
			},
			Down: []string{
				undoSoftDeleteEmptyRepositoriesBatchQuery(100001, 110000),
			},
		},
		PostDeployment: false,
	}

	allMigrations = append(allMigrations, m)
}
