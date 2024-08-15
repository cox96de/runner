package monitor

import (
	"github.com/cox96de/runner/db"
)

// Service is the service that provides the monitor functionality.
type Service struct {
	db *db.Client
}

func NewService(db *db.Client) *Service {
	return &Service{db: db}
}
