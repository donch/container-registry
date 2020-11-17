package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	m := &Migration{
		Migration: &migrate.Migration{
			Id: "20201019155133_create_tags_table",
			Up: []string{
				`CREATE TABLE IF NOT EXISTS tags (
					id bigint NOT NULL GENERATED BY DEFAULT AS IDENTITY,
					repository_id bigint NOT NULL,
					manifest_id bigint NOT NULL,
					created_at timestamp WITH time zone NOT NULL DEFAULT now(),
					updated_at timestamp WITH time zone,
					name text NOT NULL,
					CONSTRAINT pk_tags PRIMARY KEY (repository_id, id),
					CONSTRAINT fk_tags_repository_id_and_manifest_id_manifests FOREIGN KEY (repository_id, manifest_id) REFERENCES manifests (repository_id, id) ON DELETE CASCADE,
					CONSTRAINT unique_tags_repository_id_and_name UNIQUE (repository_id, name),
					CONSTRAINT check_tags_name_length CHECK ((char_length(name) <= 255))
				)
				PARTITION BY HASH (repository_id)`,
				"CREATE INDEX IF NOT EXISTS index_tags_on_repository_id_and_manifest_id ON tags USING btree (repository_id, manifest_id)",
			},
			Down: []string{
				"DROP INDEX IF EXISTS index_tags_on_repository_id_and_manifest_id CASCADE",
				"DROP TABLE IF EXISTS tags CASCADE",
			},
		},
	}

	allMigrations = append(allMigrations, m)
}
