package handler

import (
	"net/http"
	"os"
	"os/exec"

	internalmodel "github.com/cox96de/runner/internal/model"
	"github.com/cox96de/runner/util"
	"github.com/gin-gonic/gin"
	"github.com/smallnest/ringbuffer"
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
	rb := ringbuffer.New(defaultRingBufferSize)

	cmd := exec.Command(req.Command[0], req.Command[1:]...)
	cmd.Dir = req.Dir
	cmd.Stdout = rb
	cmd.Stderr = rb
	// TODO: add an option to inherit system environment variables
	cmd.Env = append(os.Environ(), util.MakeEnvPairs(req.Env)...)
	command := &command{
		Cmd:       cmd,
		logWriter: rb,
	}
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
	c.Status(http.StatusOK)
}

type command struct {
	*exec.Cmd
	logWriter *ringbuffer.RingBuffer
}
