package db

import (
	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/util"
	"gorm.io/gorm"
)

func getAllModels() []interface{} {
	return []interface{}{
		&Pipeline{}, &PipelineExecution{}, &Job{}, &JobExecution{}, &Step{}, &StepExecution{},
		&JobQueue{},
	}
}

// AutoMigrate migrates the models.
func (c *Client) AutoMigrate() error {
	return migrateModels(c.conn, getAllModels()...)
}

func migrateModels(conn *gorm.DB, models ...interface{}) error {
	err := conn.AutoMigrate(models...)
	if err != nil {
		if sqlDB, err := conn.DB(); err == nil {
			_ = sqlDB.Close()
		}
		return errors.WithMessage(err, "failed to migrate models")
	}
	return nil
}

// ToMigrateSQL generates DDL SQL for the models.
func (c *Client) ToMigrateSQL() ([]string, error) {
	return util.GenerateMigrateSQL(c.conn, getAllModels()...)
}
