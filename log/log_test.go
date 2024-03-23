package log

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"gotest.tools/v3/assert"
)

func TestParseLevel(t *testing.T) {
	level, err := ParseLevel("debug")
	assert.NilError(t, err)
	assert.Assert(t, level == Level(logrus.DebugLevel))
	_, err = ParseLevel("bad")
	assert.Error(t, err, "not a valid logrus Level: \"bad\"")
}

func TestNew(t *testing.T) {
	logger := New(&Config{
		Level:        Level(logrus.DebugLevel),
		ReportCaller: true,
	})
	buf := &bytes.Buffer{}
	logger.Logger.SetOutput(buf)
	logger.Debug("test")
	assert.Assert(t, strings.Contains(buf.String(), "test"))
	buf.Reset()
	t.Run("WithField", func(t *testing.T) {
		logger.WithField("key", "value").Debug("test")
		assert.Assert(t, strings.Contains(buf.String(), "key=value"))
		buf.Reset()
	})
	t.Run("WithFields", func(t *testing.T) {
		logger.WithFields(Fields{"key": "value"}).Debug("test")
		assert.Assert(t, strings.Contains(buf.String(), "key=value"))
	})
	t.Run("WithLogger", func(t *testing.T) {
		origin := context.Background()
		ctx := WithLogger(origin, logger)
		assert.Assert(t, ExtractLogger(ctx) == logger)
	})
}

func TestLogger_WithOutput(t *testing.T) {
	logger := ExtractLogger(context.Background())
	newLogger := logger.WithField("key", "value")
	buf := &bytes.Buffer{}
	newLogger = newLogger.WithOutput(buf)
	logger.Infof("test")
	assert.Assert(t, !strings.Contains(buf.String(), "test"))
	newLogger.Infof("test")
	assert.Assert(t, strings.Contains(buf.String(), "test"))
}
