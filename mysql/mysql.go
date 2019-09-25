package mysql

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// DB holds the connection the database
type DB struct {
	*sqlx.DB
}

func Connect() (map[string]*DB, error) {

	connections := make(map[string]*DB, 0)

	configurations := viper.GetStringMapString("db")

	for connection, dsn := range configurations {
		pool, err := sqlx.Open("mysql", dsn)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create mysql connection")
		}

		err = pool.Ping()
		if err != nil {
			return nil, errors.Wrap(err, "unable to successfully ping database")
		}

		connections[connection] = &DB{pool}
	}

	return connections, nil
}
