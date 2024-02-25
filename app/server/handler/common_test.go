package handler

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gotest.tools/v3/assert"
)

func TestJSON(t *testing.T) {
	parse, _ := time.Parse(time.RFC3339, "2021-08-01T00:00:00Z")
	a := map[string]interface{}{
		"time": timestamppb.New(parse),
		"int":  1,
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	JSON(c, 200, &a)
	assert.Assert(t, recorder.Code == 200)
	assert.Assert(t, recorder.Body.String() == "{\"int\":1,\"time\":\"2021-08-01T00:00:00Z\"}", recorder.Body.String())
}

func Test_render_Render(t *testing.T) {
	type Response struct {
		Time *timestamppb.Timestamp
	}
	engine := gin.New()

	t.Run("Render", func(t *testing.T) {
		engine.GET("/test", func(c *gin.Context) {
			parse, _ := time.Parse(time.RFC3339, "2021-08-01T00:00:00Z")
			JSON(c, 200, &Response{Time: timestamppb.New(parse)})
		})
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		engine.ServeHTTP(recorder, req)
		assert.Assert(t, recorder.Code == 200)
		assert.Assert(t, recorder.Body.String() == "{\"Time\":\"2021-08-01T00:00:00Z\"}", recorder.Body.String())
	})
	t.Run("Error", func(t *testing.T) {
		engine.GET("/error", func(c *gin.Context) {
			JSON(c, 200, &Message{Message: errors.New("some error")})
		})
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/error", nil)
		engine.ServeHTTP(recorder, req)
		assert.Assert(t, recorder.Body.String() == "{\"message\":\"some error\"}", recorder.Body.String())
	})
	t.Run("Ping", func(t *testing.T) {
		handler := &Handler{}
		engine.GET("/ping", handler.PingHandler)
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/ping", nil)
		engine.ServeHTTP(recorder, req)
		assert.Assert(t, recorder.Body.String() == "{\"message\":\"pong\"}", recorder.Body.String())
	})
}
