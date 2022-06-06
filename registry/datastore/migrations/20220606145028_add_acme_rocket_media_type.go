package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	m := &Migration{
		Migration: &migrate.Migration{
			Id: "20220606145028_add_acme_rocket_media_type",
			Up: []string{
				`INSERT INTO media_types (media_type)
					VALUES ('application/vnd.acme.rocket.docs.layer.v1+tar')
				EXCEPT
				SELECT
					media_type
				FROM
					media_types`,
			},
			Down: []string{
				`DELETE FROM media_types
					WHERE media_type = 'application/vnd.acme.rocket.docs.layer.v1+tar'`,
			},
		},
		PostDeployment: false,
	}

	allMigrations = append(allMigrations, m)
}
