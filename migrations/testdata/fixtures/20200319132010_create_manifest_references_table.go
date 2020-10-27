// +build integration

package migrationfixtures

import (
	"github.com/docker/distribution/migrations"
	migrate "github.com/rubenv/sql-migrate"
)

func init() {
	m := &migrations.Migration{Migration: &migrate.Migration{
		Id: "20200319132010_create_manifest_references_test_table",
		Up: []string{
			`CREATE TABLE IF NOT EXISTS manifest_references_test (
                id bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
                parent_id bigint NOT NULL,
                child_id bigint NOT NULL,
                created_at timestamp WITH time zone NOT NULL DEFAULT now(),
                CONSTRAINT pk_manifest_references_test PRIMARY KEY (id),
                CONSTRAINT fk_manifest_references_test_parent_id_manifests_test FOREIGN KEY (parent_id) REFERENCES manifests_test (id) ON DELETE CASCADE,
                CONSTRAINT fk_manifest_references_test_child_id_manifests_test FOREIGN KEY (child_id) REFERENCES manifests_test (id) ON DELETE CASCADE,
                CONSTRAINT uq_manifest_references_test_parent_id_child_id UNIQUE (parent_id, child_id),
				CONSTRAINT ck_manifest_references_test_parent_id_child_id_differ CHECK ((parent_id <> child_id))
            )`,
			"CREATE INDEX IF NOT EXISTS ix_manifest_references_test_parent_id ON manifest_references_test (parent_id)",
			"CREATE INDEX IF NOT EXISTS ix_manifest_references_test_child_id ON manifest_references_test (child_id)",
		},
		Down: []string{
			"DROP INDEX IF EXISTS ix_manifest_references_test_child_id CASCADE",
			"DROP INDEX IF EXISTS ix_manifest_references_test_parent_id CASCADE",
			"DROP TABLE IF EXISTS manifest_references_test CASCADE",
		},
	}}

	allMigrations = append(allMigrations, m)
}
