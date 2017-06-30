package main

import (
	"database/sql"
	"github.com/GeoNet/fits/internal/cfg"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

// fitsServer implements fits.FitsServer
type fitsServer struct {
	db *sql.DB
}

// initDB initialises the database connection and prepares resources.
// The caller should call a.close() when finished to free resources.
func (s *fitsServer) init() error {
	if s.db == nil {
		p, err := cfg.PostgresEnv()
		if err != nil {
			return errors.Wrap(err, "error reading DB config from the environment vars")
		}

		// set a statement timeout to cancel any very long running DB queries.
		// Value is int milliseconds.
		// https://www.postgresql.org/docs/9.5/static/runtime-config-client.html
		s.db, err = sql.Open("postgres", p.Connection()+" statement_timeout=600000")
		if err != nil {
			return errors.Wrap(err, "error with DB config")
		}

		s.db.SetMaxIdleConns(p.MaxIdle)
		s.db.SetMaxOpenConns(p.MaxOpen)
	}

	return nil
}

func (s *fitsServer) close() {
	s.db.Close()
}
