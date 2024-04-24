package logger

import (
	"fmt"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

type Logger struct {
	log logr.Logger

	Info  func(msg string, keysAndValues ...any)
	Error func(err error, msg string, keysAndValues ...any)
}

// WithValues returns a new Logger instance with additional key/value pairs.
func (l Logger) WithValues(keysAndValues ...any) Logger {
	log := l.log.WithValues(keysAndValues...)
	return Logger{
		log:   log,
		Info:  log.Info,
		Error: log.Error,
	}
}

func (l Logger) Namespace(namespace string) Logger {
	return l.WithValues("namespace", namespace)
}

// NewLogger builds and returns its own log struct which contains
// an implementation of 'zap.SugaredLogger' library.
func NewLogger(fileLog string) Logger {
	config := zap.NewDevelopmentConfig()

	config.OutputPaths = []string{fileLog}
	zapLog, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("log cannot be initialized (%v)?", err))
	}

	log := zapr.NewLogger(zapLog).WithValues("app", "gcli-4postman")

	return Logger{
		log:   log,
		Info:  log.Info,
		Error: log.Error,
	}
}
