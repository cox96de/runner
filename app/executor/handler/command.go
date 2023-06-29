package handler

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/cox96de/runner/lib"
	log "github.com/sirupsen/logrus"

	internalmodel "github.com/cox96de/runner/internal/model"
	"github.com/cox96de/runner/util"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

const defaultRingBufferSize = 1 << 20

// startCommandHandler is a simple ping handler, uses to validate the executor is ready.
func (h *Handler) startCommandHandler(c *gin.Context) {
	req := &internalmodel.StartCommandRequest{}
	if err := util.BindAndValidate(c, req); err != nil {
		c.JSON(http.StatusBadRequest, &internalmodel.Message{
			Message: err.Error(),
		})
		return
	}
	rb := lib.NewRingBuffer(defaultRingBufferSize)

	cmd := exec.Command(req.Command[0], req.Command[1:]...)
	cmd.Dir = req.Dir
	cmd.Stdout = rb
	cmd.Stderr = rb
	// TODO: add an option to inherit system environment variables
	cmd.Env = append(os.Environ(), util.MakeEnvPairs(req.Env)...)
	command := newCommand(cmd, rb)
	h.commandLock.Lock()
	_, exists := h.commands[req.ID]
	if exists {
		h.commandLock.Unlock()
		c.JSON(http.StatusBadRequest, &internalmodel.Message{
			Message: "command already exists",
		})
		return
	}
	h.commands[req.ID] = command
	h.commandLock.Unlock()
	if err := command.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, &internalmodel.Message{
			Message: err.Error(),
		})
		return
	}
	go func() {
		command.Wait()
		log.Infof("command %s exited", req.ID)
	}()
	c.Status(http.StatusOK)
}

const flushLogInterval = time.Millisecond * 100

// getCommandLogHandler gets commands stdout and stderr.
func (h *Handler) getCommandLogHandler(c *gin.Context) {
	req := &internalmodel.GetCommandLogRequest{}
	if err := util.BindAndValidate(c, req); err != nil {
		c.JSON(http.StatusBadRequest, &internalmodel.Message{
			Message: err.Error(),
		})
		return
	}
	logger := log.WithFields(log.Fields{"id": req.ID})
	h.commandLock.Lock()
	cmd, exists := h.commands[req.ID]
	h.commandLock.Unlock()
	if !exists {
		c.JSON(http.StatusNotFound, &internalmodel.Message{
			Message: "command not found",
		})
		return
	}
	if cmd.Exited() {
		c.Status(http.StatusOK)
	} else {
		c.Status(http.StatusPartialContent)
	}
	c.Writer.Flush()
	bufSize := 1024
	logBuf := make([]byte, bufSize)
	c.Stream(func(wr io.Writer) bool {
		n, readErr := cmd.logWriter.Read(logBuf)
		_, err := wr.Write(logBuf[:n])
		if err != nil {
			logger.Errorf("failed to write to http writer: %v", err)
			return false
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return false
			}
			logger.Infof("failed to read from log writer: %v", readErr)
			return false
		}
		if n < bufSize {
			// Cool down a bit.
			time.Sleep(flushLogInterval)
		}
		return true
	})
}

// getCommandLogHandler gets commands stdout and stderr.
func (h *Handler) getCommandStatusHandler(c *gin.Context) {
	req := &internalmodel.GetCommandStatusRequest{}
	if err := util.BindAndValidate(c, req); err != nil {
		c.JSON(http.StatusBadRequest, &internalmodel.Message{
			Message: err.Error(),
		})
		return
	}
	h.commandLock.Lock()
	cmd, exists := h.commands[req.ID]
	h.commandLock.Unlock()
	if !exists {
		c.JSON(http.StatusNotFound, &internalmodel.Message{
			Message: "command not found",
		})
		return
	}
	if !cmd.Exited() {
		// Not completed yet.
		c.JSON(http.StatusOK, &internalmodel.GetCommandStatusResponse{
			Exit: false,
		})
		return
	}
	var waitError string
	if cmd.waitError != nil {
		waitError = cmd.waitError.Error()
	}
	processState := cmd.ProcessState
	c.JSON(http.StatusOK, &internalmodel.GetCommandStatusResponse{
		ExitCode: processState.ExitCode(),
		Exit:     processState.Exited(),
		Error:    waitError,
	})
}

type command struct {
	*exec.Cmd
	logWriter io.ReadWriteCloser
	runningCh chan struct{}
	waitError error
}

func newCommand(cmd *exec.Cmd, logWriter io.ReadWriteCloser) *command {
	return &command{Cmd: cmd, logWriter: logWriter, runningCh: make(chan struct{})}
}

// Wait waits for the command to exit.
func (c *command) Wait() {
	c.waitError = c.Cmd.Wait()
	_ = c.logWriter.Close()
	close(c.runningCh)
}

// Exited returns true if the command has exited.
func (c *command) Exited() bool {
	select {
	case _, chanOpen := <-c.runningCh:
		return !chanOpen
	default:
		return false
	}
}
