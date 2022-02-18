package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	m := &Migration{
		Migration: &migrate.Migration{
			Id: "20220217183906_soft_delete_emtpy_repositories_batch_5",
			Up: []string{
				softDeleteEmptyRepositoriesBatchQuery(120001, 130000),
			},
			Down: []string{
				undoSoftDeleteEmptyRepositoriesBatchQuery(120001, 130000),
			},
		},
		PostDeployment: false,
	}

	allMigrations = append(allMigrations, m)
}
