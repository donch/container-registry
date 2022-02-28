package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	m := &Migration{
		Migration: &migrate.Migration{
			Id: "20220228110114_add_oras_media_types",
			Up: []string{
				`INSERT INTO media_types (media_type)
					VALUES ('application/vnd.unknown.config.v1+json'), ('application/vnd.oras.config.v1+json')
				EXCEPT
				SELECT
					media_type
				FROM
					media_types`,
			},
			Down: []string{
				`DELETE FROM media_types
					WHERE media_type IN (
						'application/vnd.unknown.config.v1+json',
						'application/vnd.oras.config.v1+json'
					)`,
			},
		},
		PostDeployment: false,
	}

	allMigrations = append(allMigrations, m)
}
