package entity

import (
	"encoding/json"

	"github.com/pkg/errors"
)

type Status int8

func (s *Status) String() string {
	return s.toString()
}

func (s *Status) UnmarshalJSON(bytes []byte) error {
	var str string
	err := json.Unmarshal(bytes, &str)
	if err != nil {
		return err
	}
	status, ok := str2status[str]
	if ok {
		*s = status
		return nil
	}
	return errors.Errorf("invalid status: %s", string(bytes))
}

func (s Status) MarshalJSON() ([]byte, error) {
	toString := s.toString()
	return []byte("\"" + toString + "\""), nil
}

const (
	//nolint
	statusNotStarted Status = iota - 1
	// StatusCreated is the status of a job that has been created but not yet started.
	StatusCreated
	// StatusQueued is the status of a job that has been queued but not yet started.
	StatusQueued
	statusNotStartedEnd
)

const (
	//nolint
	statusStarted Status = iota + 25
	// StatusRunning is the status of a job that is currently running.
	StatusRunning
	// StatusCanceling is the status of a job that is currently canceling.
	StatusCanceling
	//nolint
	statusStartedEnd
)

const (
	//nolint
	statusCompleted Status = iota + 50
	// StatusFailed is the status of a job that has failed.
	StatusFailed
	// StatusSkipped is the status of a job that has been skipped (completed but not really executed).
	StatusSkipped
	// StatusSucceeded is the status of a job that has succeeded.
	StatusSucceeded
	statusCompletedEnd
)

var status2str = map[Status]string{
	StatusCreated:   "created",
	StatusQueued:    "queued",
	StatusRunning:   "running",
	StatusCanceling: "canceling",
	StatusFailed:    "failed",
	StatusSkipped:   "skipped",
	StatusSucceeded: "succeeded",
}

var str2status = map[string]Status{}

func init() {
	for k, v := range status2str {
		str2status[v] = k
	}
}

func (s Status) toString() string {
	str, ok := status2str[s]
	if ok {
		return str
	}
	return "unknown"
}

// IsCompleted returns true if the status is completed.
func (s Status) IsCompleted() bool {
	return s >= statusCompleted && s <= statusCompletedEnd
}
