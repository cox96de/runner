package agent

import (
	"context"
	"runtime"
	"testing"

	"github.com/cox96de/runner/engine/shell"
	"github.com/cox96de/runner/testtool"
	log "github.com/sirupsen/logrus"

	"github.com/cox96de/runner/entity"
	"gotest.tools/v3/assert"
)

func TestExecutor_executeJob(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skip test on windows")
	}
	e := shell.NewEngine()
	log.SetLevel(log.DebugLevel)
	gitRoot, err := testtool.GetRepositoryRoot()
	assert.NilError(t, err)
	job := &entity.Job{
		Steps: []*entity.Step{
			{
				Name:             "step1",
				Commands:         []string{"ls -alh"},
				WorkingDirectory: gitRoot,
			},
			{
				Name:             "step2",
				Commands:         []string{"pwd"},
				WorkingDirectory: gitRoot,
			},
		},
	}
	execution := NewExecution(e, job)
	err = execution.Execute(context.Background())
	assert.NilError(t, err)
}
