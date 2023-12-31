//go:build integration

package migrationfixtures

import (
	"github.com/docker/distribution/registry/datastore/migrations"
	migrate "github.com/rubenv/sql-migrate"
)

func init() {
	m := &migrations.Migration{Migration: &migrate.Migration{
		Id: "20200319132237_create_tags_test_table",
		Up: []string{
			`CREATE TABLE IF NOT EXISTS tags_test (
                id bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
                repository_id bigint NOT NULL,
                manifest_id bigint NOT NULL,
                created_at timestamp WITH time zone NOT NULL DEFAULT now(),
                updated_at timestamp WITH time zone,
                name text NOT NULL,
                CONSTRAINT pk_tags_test PRIMARY KEY (id),
                CONSTRAINT fk_tags_test_repository_id_repositories FOREIGN KEY (repository_id) REFERENCES repositories_test (id) ON DELETE CASCADE,
                CONSTRAINT fk_tags_test_manifest_id_manifests FOREIGN KEY (manifest_id) REFERENCES manifests_test (id) ON DELETE CASCADE,
                CONSTRAINT unique_tags_test_name_repository_id UNIQUE (name, repository_id),
                CONSTRAINT check_tags_test_name_length CHECK ((char_length(name) <= 255))
            )`,
			"CREATE INDEX IF NOT EXISTS index_tags_test_repository_id ON tags_test (repository_id)",
			"CREATE INDEX IF NOT EXISTS index_tags_test_manifest_id ON tags_test (manifest_id)",
		},
		Down: []string{
			"DROP INDEX IF EXISTS index_tags_test_manifest_id CASCADE",
			"DROP INDEX IF EXISTS index_tags_test_repository_id CASCADE",
			"DROP TABLE IF EXISTS tags CASCADE",
		},
	}}

	allMigrations = append(allMigrations, m)
}
