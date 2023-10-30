package logutils

import (
	"net/http"

	"go.uber.org/zap"
)

func RequestWithLogger(parent *http.Request, logger *zap.Logger) *http.Request {
	return parent.WithContext(
		ContextWithLogger(parent.Context(), logger),
	)
}

func LoggerFromRequest(request *http.Request) *zap.Logger {
	return LoggerFromContext(request.Context())
}
