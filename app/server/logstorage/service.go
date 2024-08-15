package logstorage

import (
	"context"
	"time"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/external/redis"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

// TODO: it should be configurable.
// It should be greater than the timeout of a job execution.
const cacheExpireTime = time.Hour

type Service struct {
	redis   *redis.Client
	oss     OSS
	baseDir string
}

func NewService(redis *redis.Client, oss OSS) *Service {
	return &Service{redis: redis, oss: oss}
}

// Append appends log lines to the log storage.
// logName is the name of the log. For step, it is the name of the step. But there are some special log names, and generated by the system.
// These special log names typically start with _ and mainly for system logs.
func (s *Service) Append(ctx context.Context, jobExecutionID int64, logName string,
	lines []*api.LogLine,
) error {
	logSetRedisKey := buildLogSetRedisKey(jobExecutionID)
	_, err := s.redis.SAdd(ctx, logSetRedisKey, logName).Result()
	if err != nil {
		return errors.WithMessage(err, "failed to add log name to set")
	}
	encodedLines := make([]interface{}, 0, len(lines))
	for _, line := range lines {
		marshal, err := proto.Marshal(line)
		if err != nil {
			return errors.WithMessage(err, "failed to encode log line")
		}
		encodedLines = append(encodedLines, marshal)
	}
	logRedisKey := buildLogRedisKey(jobExecutionID, logName)
	_, err = s.redis.RPush(ctx, logRedisKey, encodedLines...).Result()
	if err != nil {
		return errors.WithMessage(err, "failed to push log lines to cache")
	}
	_, err = s.redis.Expire(ctx, logRedisKey, cacheExpireTime).Result()
	if err != nil {
		return errors.WithMessage(err, "failed to set expire time for cache")
	}
	// TODO: check count
	return nil
}

func (s *Service) getLogNameSet(ctx context.Context, jobExecutionID int64) ([]string, error) {
	logSetRedisKey := buildLogSetRedisKey(jobExecutionID)
	logNames, err := s.redis.SMembers(ctx, logSetRedisKey).Result()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to add log name to set")
	}
	return logNames, nil
}

// GetLogLines returns the log lines for a job execution.
// start is the start index of the log lines. It is 0-based.
// limit is the maximum number of log lines to return.
func (s *Service) GetLogLines(ctx context.Context, jobExecutionID int64, logName string, start int64,
	limit int64,
) ([]*api.LogLine, error) {
	logs, err := s.getLogsFromRedis(ctx, jobExecutionID, logName, start, limit)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get logs from cache")
	}
	if len(logs) == 0 {
		logs, err = s.getLogsFromOSS(ctx, jobExecutionID, logName, start, limit)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil, nil
			}
			return nil, errors.WithMessage(err, "failed to get logs from oss")
		}
	}
	return logs, nil
}
