package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	// Related to https://gitlab.com/gitlab-org/container-registry/-/issues/570.
	m := &Migration{
		Migration: &migrate.Migration{
			Id: "20220216124355_soft_delete_emtpy_repositories_batch_2",
			Up: []string{
				softDeleteEmptyRepositoriesBatchQuery(10001, 20000),
			},
			Down: []string{
				undoSoftDeleteEmptyRepositoriesBatchQuery(10001, 20000),
			},
		},
		PostDeployment: false,
	}

	allMigrations = append(allMigrations, m)
}
