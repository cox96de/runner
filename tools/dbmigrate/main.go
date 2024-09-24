package main

import (
	"fmt"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/db"
	"github.com/cox96de/runner/log"
	"github.com/spf13/pflag"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func main() {
	var (
		dsn         string
		dialect     string
		tablePrefix string
	)
	flagSet := pflag.NewFlagSet("dbmigrate", pflag.ExitOnError)
	flagSet.StringVar(&dsn, "dsn", "", "database connection string")
	flagSet.StringVar(&dialect, "dialect", "mysql", "database dialect, mysql, postgres, sqlite")
	flagSet.StringVar(&tablePrefix, "table-prefix", "", "table prefix")
	flagSet.Usage = func() {
		fmt.Println(`dbmigrate is a tool to generate migrate database schema.
Usage:
	dbmigrate -dsn <dsn> -dialect <dialect> -table-prefix <table-prefix>`)
		fmt.Println()
		flagSet.PrintDefaults()
	}
	err := flagSet.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	err = migrate(dialect, dsn, tablePrefix)
	if err != nil {
		log.Fatal(err)
	}
}

func migrate(dialect string, dsn string, tablePrefix string) error {
	db, err := ComposeDB(dialect, dsn, tablePrefix)
	if err != nil {
		return err
	}
	err = db.AutoMigrate()
	if err != nil {
		return err
	}
	return nil
}

func ComposeDB(dialect string, dsn string, tablePrefix string) (*db.Client, error) {
	var (
		conn *gorm.DB
		err  error
	)
	opts := &gorm.Config{
		DryRun: true,
	}
	if tablePrefix != "" {
		opts.NamingStrategy = &schema.NamingStrategy{
			TablePrefix: tablePrefix,
		}
	}
	switch db.Dialect(dialect) {
	case db.Mysql:
		conn, err = gorm.Open(mysql.Open(dsn), opts)
	case db.Postgres:
		conn, err = gorm.Open(postgres.Open(dsn), opts)
	case db.SQLite:
		conn, err = gorm.Open(sqlite.Open(dsn), opts)
	default:
		return nil, errors.Errorf("unsupported dialect: %s", dialect)
	}
	if err != nil {
		return nil, errors.WithMessage(err, "failed to open database connection")
	}
	return db.NewClient(db.Dialect(dialect), conn), nil
}
