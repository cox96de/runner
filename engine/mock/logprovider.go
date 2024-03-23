package mock

import (
	"context"
	"io"

	"github.com/cox96de/runner/engine"
)

func NewNopLogProvider() engine.LogProvider {
	return &logProvider{}
}

type logProvider struct{}

func (l *logProvider) CreateLogWriter(ctx context.Context, logName string) io.WriteCloser {
	return &nopCloser{Writer: io.Discard}
}

func (l *logProvider) GetDefaultLogWriter(ctx context.Context) io.WriteCloser {
	return &nopCloser{Writer: io.Discard}
}

type nopCloser struct {
	io.Writer
}

func (n *nopCloser) Close() error {
	return nil
}
