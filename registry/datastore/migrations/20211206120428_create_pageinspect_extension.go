package migrations

import migrate "github.com/rubenv/sql-migrate"

func init() {
	m := &Migration{
		Migration: &migrate.Migration{
			Id: "20211206120428_create_pageinspect_extension",
			Up: []string{
				`CREATE EXTENSION IF NOT EXISTS pageinspect`,
			},
			Down: []string{
				`DROP EXTENSION IF EXISTS pageinspect`,
			},
		},
		PostDeployment: false,
	}

	allMigrations = append(allMigrations, m)
}
