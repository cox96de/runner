package logstorage

import (
	"context"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/cox96de/runner/api"
	"google.golang.org/protobuf/proto"
)

func (s *Service) getLogsFromRedis(ctx context.Context, jobExecutionID int64, logName string,
	start int64, limit int64,
) ([]*api.LogLine, error) {
	end := start + limit - 1
	if limit < 0 {
		end = -1
	}
	result, err := s.redis.LRange(ctx, buildLogRedisKey(jobExecutionID, logName), start, end).Result()
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get logs from cache")
	}
	logs := make([]*api.LogLine, 0, len(result))
	for _, item := range result {
		log := new(api.LogLine)
		if err := proto.Unmarshal([]byte(item), log); err != nil {
			return nil, errors.WithMessage(err, "failed to decode log line")
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func buildLogRedisKey(jobExecutionID int64, logName string) string {
	return "log:" + strconv.FormatInt(jobExecutionID, 10) + ":" + logName
}

func buildLogSetRedisKey(jobExecutionID int64) string {
	return "log_set:" + strconv.FormatInt(jobExecutionID, 10)
}
