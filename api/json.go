package api

import (
	"encoding/json"

	"github.com/pkg/errors"
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
	toString := s.toString()
	return []byte("\"" + toString + "\""), nil
}

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
	return s >= StatusFailed && s <= StatusSucceeded
}
