package handler

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/bytedance/go-tagexpr/v2/binding"
	"github.com/cox96de/runner/log"
	"github.com/gin-gonic/gin"
)

// Bind binds the request to the object.
func Bind(c *gin.Context, obj interface{}) error {
	return binding.Bind(obj, c.Request, c.Params)
}

func getGinHandler[R any, P any](f func(ctx context.Context, request *R) (*P, error)) func(c *gin.Context) {
	return func(c *gin.Context) {
		var request R
		if err := Bind(c, &request); err != nil {
			JSON(c, http.StatusBadRequest, &Message{Message: err})
			return
		}
		response, err := f(c.Copy(), &request)
		if err != nil {
			log.ExtractLogger(c).Errorf("failed to handle request: %+v", err)
			statusCode := http.StatusInternalServerError
			var httpError *HTTPError
			if ok := errors.As(err, &httpError); ok {
				statusCode = httpError.Code
			}
			JSON(c, statusCode, &Message{Message: err})
			return
		}
		JSON(c, http.StatusOK, response)
	}
}

type HTTPError struct {
	Code       int
	CauseError error
}

func (e *HTTPError) Error() string {
	return e.CauseError.Error()
}

func (e *HTTPError) Cause() error {
	return e.CauseError
}
