package util

import (
	"context"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type recorderLogger struct {
	logger.Interface
	Statements []string
}

func (r *recorderLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, _ := fc()
	r.Statements = append(r.Statements, sql)
}

// GenerateMigrateSQL generates DDL SQL for the models.
func GenerateMigrateSQL(conn *gorm.DB, models ...interface{}) ([]string, error) {
	l := &recorderLogger{
		Interface: logger.Default.LogMode(logger.Silent),
	}
	session := conn.Session(&gorm.Session{DryRun: true, Logger: l})
	migrator := session.Migrator()
	err := migrator.AutoMigrate(models...)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to migrate")
	}
	filter := lo.Filter(l.Statements, func(item string, _ int) bool {
		return strings.HasPrefix(item, "CREATE") || strings.HasPrefix(item, "ALTER")
	})
	return filter, nil
}
