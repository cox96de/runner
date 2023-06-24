package main

import (
	"context"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"testing"

	"github.com/cox96de/runner/util"
	"github.com/gin-gonic/gin"
	"gotest.tools/v3/assert"
	"gotest.tools/v3/fs"
)

func Test_loadConfig(t *testing.T) {
	type args struct {
		arguments []string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name: "error",
			args: args{
				arguments: []string{"--port", "abc"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "default",
			args: args{
				arguments: []string{},
			},
			want:    &Config{Port: 8080},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadConfig(tt.args.arguments)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_composeListener(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		listener, err := composeListener(&Config{})
		assert.ErrorContains(t, err, "no listener configured")
		assert.Assert(t, listener == nil)
	})
	t.Run("tcp_port", func(t *testing.T) {
		randomPort := util.RandomInt(1024, 65535)
		listener, err := composeListener(&Config{Port: int(randomPort)})
		assert.NilError(nil, err)
		engine := gin.New()
		engine.Any("/ping", func(context *gin.Context) {})
		go func() {
			err := engine.RunListener(listener)
			checkError(err)
		}()
		resp, err := http.Get("http://127.0.0.1:" + strconv.Itoa(int(randomPort)) + "/ping")
		assert.NilError(t, err)
		assert.DeepEqual(t, resp.StatusCode, http.StatusOK)
	})
	t.Run("unix_socket", func(t *testing.T) {
		testDir := fs.NewDir(t, "tmp")
		socketPath := testDir.Join("test.sock")
		listener, err := composeListener(&Config{SocketPath: socketPath})
		assert.NilError(t, err)
		engine := gin.New()
		engine.Any("/ping", func(context *gin.Context) {})
		go func() {
			err := engine.RunListener(listener)
			checkError(err)
		}()
		request, err := http.NewRequest("GET", "http://unixsocket/ping", nil)
		assert.NilError(t, err)
		client := &http.Client{
			Transport: &http.Transport{
				DialContext: func(context context.Context, network, addr string) (net.Conn, error) {
					conn, err := net.Dial("unix", socketPath)
					return conn, err
				},
			},
		}
		resp, err := client.Do(request)
		assert.NilError(t, err)
		assert.DeepEqual(t, resp.StatusCode, http.StatusOK)
	})
}
