package db

import (
	"context"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

type recorderLogger struct {
	logger.Interface
	Statements []string
}

func (r *recorderLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, _ := fc()
	r.Statements = append(r.Statements, sql)
}

// ToMigrateSQL generates DDL SQL for the models.
func (c *Client) ToMigrateSQL() ([]string, error) {
	l := &recorderLogger{
		Interface: logger.Default.LogMode(logger.Silent),
	}
	session := c.conn.Session(&gorm.Session{DryRun: true, Logger: l})
	migrator := session.Migrator()
	err := migrator.AutoMigrate(getAllModels()...)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to migrate")
	}
	filter := lo.Filter(l.Statements, func(item string, _ int) bool {
		return strings.HasPrefix(item, "CREATE") || strings.HasPrefix(item, "ALTER")
	})
	return filter, nil
}
