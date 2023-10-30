package httplogger

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/flashbots/node-healthchecker/logutils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func Middleware(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate request ID (`base64` to shorten its string representation)
		_uuid := [16]byte(uuid.New())
		httpRequestID := base64.RawStdEncoding.EncodeToString(_uuid[:])

		l := logger.With(
			zap.String("httpRequestID", httpRequestID),
			zap.String("logType", "activity"),
		)
		r = logutils.RequestWithLogger(r, l)

		// Handle panics
		defer func() {
			if msg := recover(); msg != nil {
				w.WriteHeader(http.StatusInternalServerError)
				var method, url string
				if r != nil {
					method = r.Method
					url = r.URL.EscapedPath()
				}
				l.Error("HTTP request handler panicked",
					zap.Any("error", msg),
					zap.String("method", method),
					zap.String("url", url),
				)
			}
		}()

		start := time.Now()
		wrapped := wrapResponseWriter(w)
		next.ServeHTTP(wrapped, r)

		// Passing request stats both in-message (for the human reader)
		// as well as inside the structured log (for the machine parser)
		logger.Info(fmt.Sprintf("%s: %s %s %d", r.URL.Scheme, r.Method, r.URL.EscapedPath(), wrapped.Status()),
			zap.Int("durationMs", int(time.Since(start).Milliseconds())),
			zap.Int("status", wrapped.Status()),
			zap.String("httpRequestID", httpRequestID),
			zap.String("logType", "access"),
			zap.String("method", r.Method),
			zap.String("path", r.URL.EscapedPath()),
			zap.String("userAgent", r.Header.Get("user-agent")),
		)
	})
}
