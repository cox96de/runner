package agent

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/log"
)

// logCollector reads log from executor or other source and upload to server.
type logCollector struct {
	jobExecution *api.JobExecution
	// logger is used to log error.
	logger  *log.Logger
	client  api.ServerClient
	logName string
	// lineNo is the line number of the log. Line number starts from 0.
	lineNo int64
	// start is the time when the logCollector is created.
	// It is used to calculate the timestamp of the log.
	start time.Time
	// logs is the buffer of logs.
	logs []*api.LogLine
	// incompleteBytes is the buffer of incomplete line.
	incompleteBytes []byte
	lock            sync.RWMutex
	// notify is used to notify the work routine to flush logs.
	notify chan struct{}
	// close is used to notify the work routine to stop.
	close chan struct{}
	// worker indicates the work routine is running.
	worker        chan struct{}
	flushInterval time.Duration
}

func newLogCollector(client api.ServerClient, jobExecution *api.JobExecution, logName string, logger *log.Logger,
	flushInterval time.Duration,
) *logCollector {
	l := &logCollector{
		client:        client,
		jobExecution:  jobExecution,
		logName:       logName,
		logger:        logger,
		start:         time.Now(),
		flushInterval: flushInterval,
		notify:        make(chan struct{}, 1),
		close:         make(chan struct{}),
		worker:        make(chan struct{}),
	}
	go l.work()
	return l
}

func (l *logCollector) Write(p []byte) (n int, err error) {
	l.lock.Lock()
	defer l.lock.Unlock()
	scanner := bufio.NewScanner(bytes.NewReader(append(l.incompleteBytes, p...)))
	buf := &bytes.Buffer{}
	scanner.Split(getLineScanner(buf))
	t := time.Since(l.start).Seconds()
	for scanner.Scan() {
		line := scanner.Text()
		l.lineNo++
		l.logs = append(l.logs, &api.LogLine{
			Timestamp: int64(t),
			Number:    l.lineNo,
			Output:    line,
		})
	}
	l.incompleteBytes = buf.Bytes()
	l.notifyWrite()
	return len(p), nil
}

func (l *logCollector) Close() error {
	if l.incompleteBytes != nil {
		_, _ = l.Write([]byte("\n"))
	}
	select {
	case <-l.close:
		return errors.New("log already closed")
	default:
	}
	close(l.close)
	var err error
	select {
	// Wait for the worker to stop.
	// In worker, the flush is called, concurrent flush is not allowed, it might cause data race.
	case <-l.worker:
	case <-time.After(time.Second):
		return errors.WithMessage(err, "background worker is not stopped in time")
	}
	// `worker` is closed, no need to lock.
	if len(l.logs) == 0 {
		return nil
	}
	for i := 0; i < 3; i++ {
		if err = l.flush(); err == nil {
			return nil
		}
	}
	return errors.WithMessage(err, "failed to flush log")
}

func (l *logCollector) work() {
	defer close(l.worker)
	for {
		select {
		case <-l.notify:
			select {
			case <-time.After(l.flushInterval):
				if err := l.flush(); err != nil {
					l.logger.Warnf("failed to flush log: %v", err)
				}
			case <-l.close:
				return
			}
		case <-l.close:
			return
		}
	}
}

func (l *logCollector) notifyWrite() {
	select {
	case l.notify <- struct{}{}:
	default:
	}
}

// flush flushes logs to server.
// Notice: it is not thread safe. It should be called in a serial way.
func (l *logCollector) flush() error {
	l.lock.RLock()
	size := len(l.logs)
	logs := l.logs[:size]
	l.lock.RUnlock()
	if size == 0 {
		return nil
	}
	// TODO: use context with timeout.
	_, err := l.client.UploadLogLines(context.Background(), &api.UpdateLogLinesRequest{
		JobID:          l.jobExecution.JobID,
		JobExecutionID: l.jobExecution.ID,
		Name:           l.logName,
		Lines:          logs,
	})
	if err == nil {
		l.lock.Lock()
		l.logs = l.logs[size:]
		l.lock.Unlock()
		return nil
	}
	l.logger.Warnf("failed to upload log: %v", err)
	return err
}

func getLineScanner(incompleteLine io.Writer) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		// \r is priority to \n.
		rIdx := bytes.IndexByte(data, '\r')
		nIdx := bytes.IndexByte(data, '\n')
		switch {
		case rIdx >= 0 && rIdx+1 == nIdx:
			// We met \r\n.
			return nIdx + 1, dropCR(data[0:nIdx]), nil
		case rIdx >= 0 && (rIdx < nIdx || nIdx < 0):
			// \r is in front of \n but not \r\n or no \n.
			return rIdx + 1, data[0:rIdx], nil
		case nIdx >= 0:
			// We have a full newline-terminated line.
			return nIdx + 1, data[0:nIdx], nil
		case atEOF:
			// We met an incomplete line.
			_, err := incompleteLine.Write(data)
			return len(data), nil, err
		default:
			// Request more data.
			return 0, nil, nil
		}
	}
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
