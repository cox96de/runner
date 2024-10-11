package monitor

import (
	"github.com/cox96de/runner/app/server/dispatch"
	"github.com/cox96de/runner/app/server/eventhook"
	"github.com/cox96de/runner/app/server/logstorage"
	"github.com/cox96de/runner/db"
)

// Service is the service that provides the monitor functionality.
type Service struct {
	db                *db.Client
	logstorageService *logstorage.Service
	eventhook         *eventhook.Service
	dispatchService   *dispatch.Service
}

func NewService(db *db.Client, logstorageService *logstorage.Service, eventhook *eventhook.Service, dispatchService *dispatch.Service) *Service {
	return &Service{db: db, logstorageService: logstorageService, eventhook: eventhook, dispatchService: dispatchService}
}
