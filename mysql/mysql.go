package mysql

import (
	"log"

	"github.com/ddouglas/monocle"
	"github.com/jmoiron/sqlx"
	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
)

// DB holds the connection the database
type DB struct {
	*sqlx.DB
}

func Connect() (*DB, error) {

	var config monocle.Config
	err := envconfig.Process("", &config)
	if err != nil {
		log.Fatal("unable initialize environment variables")
	}

	pool, err := sqlx.Open("mysql", config.DBDsn)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create mysql connection")
	}

	err = pool.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "unable to successfully ping database")
	}

	return &DB{pool}, nil
}
