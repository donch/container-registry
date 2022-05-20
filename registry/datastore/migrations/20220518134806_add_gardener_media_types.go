package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	m := &Migration{
		Migration: &migrate.Migration{
			Id: "20220518134806_add_gardener_media_types",
			Up: []string{
				`INSERT INTO media_types (media_type)
					VALUES ('application/vnd.gardener.cloud.cnudie.component-descriptor.v2+yaml+tar'), ('application/vnd.gardener.landscaper.blueprint.v1+tar+gzip')
				EXCEPT
				SELECT
					media_type
				FROM
					media_types`,
			},
			Down: []string{
				`DELETE FROM media_types
					WHERE media_type IN (
						'application/vnd.gardener.cloud.cnudie.component-descriptor.v2+yaml+tar',
						'application/vnd.gardener.landscaper.blueprint.v1+tar+gzip'
					)`,
			},
		},
		PostDeployment: false,
	}

	allMigrations = append(allMigrations, m)
}
