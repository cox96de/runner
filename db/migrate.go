package db

import (
	"github.com/cockroachdb/errors"
	"gorm.io/gorm"
)

// AutoMigrate migrates the models.
func (c *Client) AutoMigrate() error {
	return migrateModels(c.conn, &Pipeline{}, &PipelineExecution{}, &Job{}, &JobExecution{}, &Step{}, &StepExecution{},
		&JobQueue{})
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
