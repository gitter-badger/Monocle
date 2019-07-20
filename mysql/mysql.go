package mysql

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// DB holds the connection the database
type DB struct {
	*sqlx.DB
}

func Connect(dsn string) (db *DB, err error) {
	pool, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create mysql connection")
	}

	err = pool.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "unable to successfully ping database")
	}

	return &DB{pool}, nil
}
