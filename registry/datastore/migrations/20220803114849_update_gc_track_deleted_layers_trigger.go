package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	m := &Migration{
		Migration: &migrate.Migration{
			Id: "20220803114849_update_gc_track_deleted_layers_trigger",
			// switch from row-level to statement-level execution
			Up: []string{
				"DROP TRIGGER IF EXISTS gc_track_deleted_layers_trigger ON layers",
				`CREATE TRIGGER gc_track_deleted_layers_trigger
					AFTER DELETE ON layers REFERENCING OLD TABLE AS old_table
					FOR EACH STATEMENT
					EXECUTE FUNCTION gc_track_deleted_layers ()`,
			},
			// restore previous version
			Down: []string{
				"DROP TRIGGER IF EXISTS gc_track_deleted_layers_trigger ON layers",
				`CREATE TRIGGER gc_track_deleted_layers_trigger
					AFTER DELETE ON layers
					FOR EACH ROW
					EXECUTE FUNCTION gc_track_deleted_layers ()`,
			},
		},
		PostDeployment: false,
	}

	allMigrations = append(allMigrations, m)
}
