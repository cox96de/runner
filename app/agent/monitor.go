package agent

import (
	"context"
	"time"

	"github.com/cox96de/runner/api"
	"github.com/cox96de/runner/log"
)

const maxHeartbeatTimeout = time.Second * 30

func (e *Execution) startMonitor(ctx context.Context) {
	e.jobCtx, e.jobCanceller = context.WithCancel(ctx)
	go e.monitorJobTimeout(ctx)
	go e.monitorHeartbeat(ctx, time.Second*5, maxHeartbeatTimeout)
}

func (e *Execution) monitorJobTimeout(ctx context.Context) {
	if e.job.Timeout <= 0 {
		return
	}
	select {
	case <-time.After(time.Duration(e.job.Timeout) * time.Second):
		e.abortedReason.Store(uint32(TimeoutAbortReason))
		e.jobCanceller()
	case <-ctx.Done():
	}
}

func (e *Execution) monitorHeartbeat(ctx context.Context, internal time.Duration, timeout time.Duration) {
	logger := log.ExtractLogger(ctx)
	ticker := time.NewTicker(internal)
	defer ticker.Stop()
	lastHeartbeat := time.Now()
	for {
		select {
		case <-ticker.C:
			_, err := e.client.Heartbeat(ctx, &api.HeartbeatRequest{JobExecutionID: e.jobExecution.ID})
			if err != nil {
				logger.Error("failed to send heartbeat", err)
				if time.Since(lastHeartbeat) > timeout {
					logger.Errorf("failed to send heartbeat for %d seconds, stopping execution", maxHeartbeatTimeout/time.Second)
					e.jobCanceller()
				}
				// TODO: abort early when job is already completed.
				continue
			}
			lastHeartbeat = time.Now()
		case <-ctx.Done():
			return
		}
	}
}
