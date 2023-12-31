//go:build integration

package migrationfixtures

import (
	"github.com/docker/distribution/registry/datastore/migrations"
	migrate "github.com/rubenv/sql-migrate"
)

func init() {
	m := &migrations.Migration{Migration: &migrate.Migration{
		Id: "20200319122755_create_repositories_test_table",
		Up: []string{
			`CREATE TABLE IF NOT EXISTS repositories_test (
                id bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
                parent_id bigint,
                created_at timestamp WITH time zone NOT NULL DEFAULT now(),
    			updated_at timestamp WITH time zone,
                name text NOT NULL,
                path text NOT NULL,
                CONSTRAINT pk_repositories_test PRIMARY KEY (id),
                CONSTRAINT fk_repositories_test_parent_id_repositories FOREIGN KEY (parent_id) REFERENCES repositories_test (id) ON DELETE CASCADE,
                CONSTRAINT unique_repositories_test_path UNIQUE (path),
                CONSTRAINT check_repositories_test_name_length CHECK ((char_length(name) <= 255)),
                CONSTRAINT check_repositories_test_path_length CHECK ((char_length(path) <= 255))
            )`,
			"CREATE INDEX IF NOT EXISTS index_repositories_parent_id ON repositories_test (parent_id)",
		},
		Down: []string{
			"DROP INDEX IF EXISTS index_repositories_test_parent_id CASCADE",
			"DROP TABLE IF EXISTS repositories_test CASCADE",
		},
	}}

	allMigrations = append(allMigrations, m)
}
