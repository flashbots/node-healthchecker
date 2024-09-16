package logutils

import (
	"errors"
	"log"
	"strings"

	"go.uber.org/zap"
)

type httpServerErrorLogger struct {
	logger *zap.Logger
}

func (s *httpServerErrorLogger) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	s.logger.Warn("HTTP server encountered an error",
		zap.Error(errors.New(msg)),
	)
	return len(p), nil
}

func NewHttpServerErrorLogger(logger *zap.Logger) *log.Logger {
	wrapped := &httpServerErrorLogger{
		logger: logger,
	}
	return log.New(wrapped, "", 0)
}
