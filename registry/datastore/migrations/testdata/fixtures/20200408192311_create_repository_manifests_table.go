//go:build integration

package migrationfixtures

import (
	"github.com/docker/distribution/registry/datastore/migrations"
	migrate "github.com/rubenv/sql-migrate"
)

func init() {
	m := &migrations.Migration{Migration: &migrate.Migration{
		Id: "20200408192311_create_repository_manifests_test_table",
		Up: []string{
			`CREATE TABLE IF NOT EXISTS repository_manifests_test (
                id bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
                repository_id bigint NOT NULL,
                manifest_id bigint NOT NULL,
                created_at timestamp WITH time zone NOT NULL DEFAULT now(),
                CONSTRAINT pk_repository_manifests_test PRIMARY KEY (id),
                CONSTRAINT fk_repository_manifests_test_repository_id_repositories FOREIGN KEY (repository_id) REFERENCES repositories_test (id) ON DELETE CASCADE,
                CONSTRAINT fk_repository_manifests_test_manifest_id_manifests FOREIGN KEY (manifest_id) REFERENCES manifests_test (id) ON DELETE CASCADE,
                CONSTRAINT unique_repository_manifests_test_repository_id_manifest_id UNIQUE (repository_id, manifest_id)
            )`,
			"CREATE INDEX IF NOT EXISTS index_repository_manifests_test_repository_id ON repository_manifests_test (repository_id)",
			"CREATE INDEX IF NOT EXISTS index_repository_manifests_test_manifest_id ON repository_manifests_test (manifest_id)",
		},
		Down: []string{
			"DROP INDEX IF EXISTS index_repository_manifests_test_manifest_id CASCADE",
			"DROP INDEX IF EXISTS index_repository_manifests_test_repository_id CASCADE",
			"DROP TABLE IF EXISTS repository_manifests_test CASCADE",
		},
	}}

	allMigrations = append(allMigrations, m)
}
