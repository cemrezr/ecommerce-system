package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

func Connect(dsn string, log zerolog.Logger) *sqlx.DB {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatal().
			Err(err).
			Str("dsn", dsn).
			Msg("Failed to connect to PostgreSQL")
	}

	log.Info().Str("component", "postgres").Msg("Connected to PostgreSQL")
	return db
}
