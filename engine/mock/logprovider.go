package mock

import (
	"context"
	"io"

	"github.com/cox96de/runner/engine"
	"github.com/cox96de/runner/util"
)

func NewNopLogProvider() engine.LogProvider {
	return &logProvider{}
}

type logProvider struct{}

func (l *logProvider) CreateLogWriter(ctx context.Context, logName string) io.WriteCloser {
	return util.NopCloser(io.Discard)
}

func (l *logProvider) GetDefaultLogWriter(ctx context.Context) io.WriteCloser {
	return util.NopCloser(io.Discard)
}
