package monitor

import (
	"github.com/cox96de/runner/app/server/logstorage"
	"github.com/cox96de/runner/db"
)

// Service is the service that provides the monitor functionality.
type Service struct {
	db                *db.Client
	logstorageService *logstorage.Service
}

func NewService(db *db.Client, logstorageService *logstorage.Service) *Service {
	return &Service{db: db, logstorageService: logstorageService}
}
