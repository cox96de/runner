package api

import (
	"encoding/json"

	"github.com/cockroachdb/errors"
)

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
	toString := s.ToString()
	return []byte("\"" + toString + "\""), nil
}

var status2str = map[Status]string{
	StatusCreated:   "created",
	StatusQueued:    "queued",
	StatusPreparing: "preparing",
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

func (s Status) ToString() string {
	str, ok := status2str[s]
	if ok {
		return str
	}
	return "unknown"
}

// IsCompleted returns true if the status is completed.
func (s Status) IsCompleted() bool {
	return s >= StatusFailed && s <= StatusSucceeded
}

// IsRunning returns true if the status is dispatched to agent.
func (s Status) IsRunning() bool {
	return s >= StatusPreparing && s < StatusFailed
}

// IsPreDispatch returns true if the status is before dispatching.
func (s Status) IsPreDispatch() bool {
	return s <= StatusQueued
}
