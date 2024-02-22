package handler

import (
	"net/http/httptest"
	"testing"
	"time"

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
