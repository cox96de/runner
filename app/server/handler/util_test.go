package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cockroachdb/errors"

	"github.com/gin-gonic/gin"
	"gotest.tools/v3/assert"
)

func TestBind(t *testing.T) {
	engine := gin.New()
	type Test struct {
		Name   string `json:"name"`
		Header string `header:"header"`
		Query  string `query:"query"`
		Path   string `path:"path"`
	}
	req := &Test{}
	engine.Any("/test/:path", func(c *gin.Context) {
		if err := Bind(c, req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	})
	server := httptest.NewServer(engine)
	defer server.Close()
	request, err := http.NewRequest(http.MethodPost, server.URL+"/test/path?query=query_value", bytes.NewReader([]byte("{\"name\":\"test\"}")))
	assert.NilError(t, err)
	request.Header.Add("header", "header_value")
	request.Header.Add("Content-Type", "application/json")
	assert.NilError(t, err)
	do, err := server.Client().Do(request)
	assert.NilError(t, err)
	assert.Equal(t, do.StatusCode, http.StatusOK)
	assert.DeepEqual(t, req, &Test{
		Name:   "test",
		Header: "header_value",
		Query:  "query_value",
		Path:   "path",
	})
}

func Test_getGinHandler(t *testing.T) {
	engine := gin.New()
	engine.Any("/ping", getGinHandler(func(ctx context.Context, request *string) (*string, error) {
		return nil, &HTTPError{Code: 444, CauseError: errors.New(t.Name())}
	}))
	server := httptest.NewServer(engine)
	request, err := http.NewRequest(http.MethodGet, server.URL+"/ping", nil)
	assert.NilError(t, err)
	do, err := server.Client().Do(request)
	assert.NilError(t, err)
	assert.Equal(t, do.StatusCode, 444)
}
