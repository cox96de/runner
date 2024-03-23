package log

import (
	"context"
	"io"

	"github.com/sirupsen/logrus"
)

var (
	loggerKey     = &struct{}{}
	defaultLogger = logrus.StandardLogger()
)

type (
	Fields logrus.Fields
	Level  logrus.Level
)

type Logger struct {
	*logrus.Entry
}

type Config struct {
	Level        Level
	ReportCaller bool
}

func ParseLevel(level string) (Level, error) {
	l, err := logrus.ParseLevel(level)
	return Level(l), err
}

func New(c *Config) *Logger {
	logger := logrus.New()
	if c.Level > 0 {
		logger.SetLevel(logrus.Level(c.Level))
	}
	if c.ReportCaller {
		logger.SetReportCaller(c.ReportCaller)
	}
	return &Logger{Entry: logrus.NewEntry(logger)}
}

func ExtractLogger(ctx context.Context) *Logger {
	l, ok := ctx.Value(loggerKey).(*Logger)
	if !ok {
		return &Logger{Entry: logrus.NewEntry(defaultLogger)}
	}
	return l
}

func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func (l *Logger) WithOutput(out io.Writer) *Logger {
	logger := &logrus.Logger{
		Out:          out,
		Hooks:        l.Logger.Hooks,
		Formatter:    l.Logger.Formatter,
		ReportCaller: l.Logger.ReportCaller,
		Level:        l.Logger.Level,
		BufferPool:   l.Logger.BufferPool,
	}
	ll := l.WithFields(nil)
	ll.Logger = logger
	return ll
}

func (l *Logger) WithFields(f Fields) *Logger {
	e := l.Entry.WithFields(logrus.Fields(f))
	return &Logger{Entry: e}
}

func (l *Logger) WithField(key string, value interface{}) *Logger {
	e := l.Entry.WithField(key, value)
	return &Logger{Entry: e}
}

func Errorf(s string, args ...interface{}) {
	defaultLogger.Errorf(s, args...)
}

func Infof(s string, args ...interface{}) {
	defaultLogger.Infof(s, args...)
}

func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

func Warningf(s string, args ...interface{}) {
	defaultLogger.Warningf(s, args...)
}

func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}
