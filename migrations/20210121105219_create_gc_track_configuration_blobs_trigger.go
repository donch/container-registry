package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	m := &Migration{
		Migration: &migrate.Migration{
			Id: "20210121105219_create_gc_track_configuration_blobs_trigger",
			Up: []string{
				`DO $$
				BEGIN
					IF NOT EXISTS (
						SELECT
							1
						FROM
							pg_trigger
						WHERE
							tgname = 'gc_track_configuration_blobs_trigger') THEN
						CREATE TRIGGER gc_track_configuration_blobs_trigger
							AFTER INSERT ON manifests
							FOR EACH ROW
							EXECUTE PROCEDURE gc_track_configuration_blobs ();
					END IF;
				END
				$$`,
			},
			Down: []string{
				"DROP TRIGGER IF EXISTS gc_track_configuration_blobs_trigger ON manifests CASCADE",
			},
		},
		PostDeployment: false,
	}

	allMigrations = append(allMigrations, m)
}
