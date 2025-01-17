package composer

import (
	"github.com/cockroachdb/errors"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type DB struct {
	// Dialect is the database dialect, support sqlite, mysql, postgres.
	Dialect     string `mapstructure:"dialect" yaml:"dialect"`
	DSN         string `mapstructure:"dsn" yaml:"dsn"`
	TablePrefix string `mapstructure:"table_prefix" yaml:"table_prefix"`
}

func ComposeDB(c *DB) (*gorm.DB, error) {
	var (
		conn    *gorm.DB
		err     error
		dialect = c.Dialect
		dsn     = c.DSN
	)
	opts := &gorm.Config{}
	if c.TablePrefix != "" {
		opts.NamingStrategy = &schema.NamingStrategy{
			TablePrefix: c.TablePrefix,
		}
	}
	switch dialect {
	case mysql.Dialector{}.Name():
		conn, err = gorm.Open(mysql.Open(dsn), opts)
	case postgres.Dialector{}.Name():
		conn, err = gorm.Open(postgres.Open(dsn), opts)
	case sqlite.Dialector{}.Name():
		conn, err = gorm.Open(sqlite.Open(dsn), opts)
	default:
		return nil, errors.Errorf("unsupported dialect: %s", dialect)
	}
	if err != nil {
		return nil, errors.WithMessage(err, "failed to open database connection")
	}
	return conn, nil
}
