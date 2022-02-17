package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	m := &Migration{
		Migration: &migrate.Migration{
			Id: "20220217095720_soft_delete_emtpy_repositories_batch_4",
			Up: []string{
				softDeleteEmptyRepositoriesBatchQuery(80001, 90000),
			},
			Down: []string{
				undoSoftDeleteEmptyRepositoriesBatchQuery(80001, 90000),
			},
		},
		PostDeployment: false,
	}

	allMigrations = append(allMigrations, m)
}
