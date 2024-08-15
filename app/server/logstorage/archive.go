package logstorage

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strconv"

	"github.com/cox96de/runner/log"

	"github.com/cox96de/runner/api"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func (s *Service) buildKey(jobExecutionID int64, logName string) string {
	return filepath.Join(s.baseDir, strconv.FormatInt(jobExecutionID, 10), logName)
}

func (s *Service) getLogsFromOSS(ctx context.Context, jobExecutionID int64, logName string,
	start int64, limit int64,
) ([]*api.LogLine, error) {
	key := s.buildKey(jobExecutionID, logName)
	object, err := s.oss.Open(ctx, key)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to get object")
	}
	defer object.Close()
	content, err := io.ReadAll(object)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to read content")
	}
	archive := api.ArchiveLogs{}
	err = proto.Unmarshal(content, &archive)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to unmarshal archive")
	}
	if start > int64(len(archive.Logs)) {
		return nil, nil
	}
	return archive.Logs[start:min(start+limit, int64(len(archive.Logs)))], nil
}

// Archive archives logs to S3.
// It should be invoked when there are no more logs to be appended.
// Typically, it is invoked after the job is finished.
func (s *Service) Archive(ctx context.Context, jobExecutionID int64) error {
	log.ExtractLogger(ctx).WithFields(log.Fields{"job_execution_id": jobExecutionID}).
		Infof("archive logs")
	logNameSet, err := s.getLogNameSet(ctx, jobExecutionID)
	if err != nil {
		return errors.WithMessage(err, "failed to get log name set")
	}
	for _, logName := range logNameSet {
		key := s.buildKey(jobExecutionID, logName)
		lines, err := s.getLogsFromRedis(ctx, jobExecutionID, logName, 0, -1)
		if err != nil {
			return errors.WithMessage(err, "failed to get logs from redis")
		}
		archive := api.ArchiveLogs{Logs: lines}
		content, err := proto.Marshal(&archive)
		if err != nil {
			return errors.WithMessage(err, "failed to marshal archive")
		}
		_, err = s.oss.Save(ctx, key, bytes.NewReader(content))
		if err != nil {
			return errors.WithMessage(err, "failed to put object")
		}
		_, err = s.redis.Del(ctx, buildLogRedisKey(jobExecutionID, logName)).Result()
		if err != nil {
			return errors.WithMessage(err, "failed to remove log from cache")
		}
		_, err = s.redis.SRem(ctx, buildLogSetRedisKey(jobExecutionID), logName).Result()
		if err != nil {
			return errors.WithMessage(err, "failed to remove log name from set")
		}
	}
	return nil
}
