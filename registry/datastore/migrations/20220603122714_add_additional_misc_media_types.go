package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	m := &Migration{
		Migration: &migrate.Migration{
			Id: "20220603122714_add_additional_misc_media_types",
			Up: []string{
				`INSERT INTO media_types (media_type)
					VALUES 
						('text/html; charset=utf-8')
				EXCEPT
				SELECT
					media_type
				FROM
					media_types`,
			},
			Down: []string{
				`DELETE FROM media_types
					WHERE media_type = 'text/html; charset=utf-8'`,
			},
		},
		PostDeployment: false,
	}

	allMigrations = append(allMigrations, m)
}
