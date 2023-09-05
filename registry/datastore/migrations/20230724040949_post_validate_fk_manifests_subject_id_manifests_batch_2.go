//go:build !integration

package migrations

import (
	"fmt"

	migrate "github.com/rubenv/sql-migrate"
)

func init() {
	var ups []string

	for i := 13; i <= 25; i++ {
		ups = append(ups, fmt.Sprintf(
			`DO $$
			BEGIN
				IF EXISTS (
					SELECT
						1
					FROM
						pg_catalog.pg_constraint
					WHERE
						conrelid = 'partitions.manifests_p_%d'::regclass
						AND conname = 'fk_manifests_subject_id_manifests'
				) THEN
					ALTER TABLE partitions.manifests_p_%d VALIDATE CONSTRAINT fk_manifests_subject_id_manifests;
				END IF;
			END;
			$$`, i, i))
	}

	m := &Migration{
		Migration: &migrate.Migration{
			Id:                   "20230724040949_post_validate_fk_manifests_subject_id_manifests_batch_2",
			Up:                   ups,
			Down:                 []string{},
			DisableTransactionUp: true,
		},
		PostDeployment: true,
	}

	allMigrations = append(allMigrations, m)
}