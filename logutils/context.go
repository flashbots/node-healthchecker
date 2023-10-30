package logutils

import (
	"context"

	"go.uber.org/zap"
)

type contextKey string

const loggerContextKey contextKey = "logger"

func ContextWithLogger(parent context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(parent, loggerContextKey, logger)
}

func LoggerFromContext(ctx context.Context) *zap.Logger {
	if l, found := ctx.Value(loggerContextKey).(*zap.Logger); found {
		return l
	}
	return zap.L()
}
