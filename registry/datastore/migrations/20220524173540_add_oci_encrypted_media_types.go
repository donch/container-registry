package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	m := &Migration{
		Migration: &migrate.Migration{
			Id: "20220524173540_add_oci_encrypted_media_types",
			Up: []string{
				`INSERT INTO media_types (media_type)
					VALUES ('application/vnd.oci.image.layer.v1.tar+gzip+encrypted'), ('application/vnd.oci.image.layer.v1.tar+encrypted')
				EXCEPT
				SELECT
					media_type
				FROM
					media_types`,
			},
			Down: []string{
				`DELETE FROM media_types
					WHERE media_type IN (
						'application/vnd.oci.image.layer.v1.tar+gzip+encrypted',
						'application/vnd.oci.image.layer.v1.tar+encrypted'
					)`,
			},
		},
		PostDeployment: false,
	}

	allMigrations = append(allMigrations, m)
}
