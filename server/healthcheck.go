package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/flashbots/node-healthchecker/healthcheck"
	"github.com/flashbots/node-healthchecker/logutils"
	"go.uber.org/zap"
)

func (s *Server) healthcheck(w http.ResponseWriter, r *http.Request) {
	l := logutils.LoggerFromRequest(r)

	count := len(s.monitors)
	results := make(chan *healthcheck.Result, count)

	for _, m := range s.monitors {
		monitor := m // https://go.dev/blog/loopvar-preview
		ctx, cancel := context.WithTimeout(r.Context(), s.cfg.Healthcheck.Timeout)
		defer cancel()
		go func() {
			results <- monitor(ctx)
		}()
	}

	errs := []error{}
	warns := []error{}
	for count > 0 {
		count--
		if res := <-results; res != nil {
			if !res.Ok {
				errs = append(errs, res.Err)
			} else if res.Err != nil {
				warns = append(warns, res.Err)
			}
		}
	}
	close(results)

	switch {
	case len(errs) == 0 && len(warns) == 0:
		w.WriteHeader(s.cfg.HttpStatus.Ok)
		return

	case len(errs) > 0:
		w.WriteHeader(s.cfg.HttpStatus.Error)
		w.Header().Set("Content-Type", "application/text")

		for idx, err := range errs {
			line := fmt.Sprintf("%d: error: %s\n", idx, err)
			_, err := w.Write([]byte(line))
			if err != nil {
				l.Error("Failed to write the response body",
					zap.Error(err),
				)
			}
		}
		offset := len(errs)
		for idx, warn := range warns {
			line := fmt.Sprintf("%d: warning: %s\n", offset+idx, warn)
			_, err := w.Write([]byte(line))
			if err != nil {
				l.Error("Failed to write the response body",
					zap.Error(err),
				)
			}
		}

		l.Warn("Healthcheck encountered upstream error(s)",
			zap.Error(errors.Join(errs...)),
			zap.Int("http_status", s.cfg.HttpStatus.Error),
		)

	case len(errs) == 0 && len(warns) > 0:
		w.WriteHeader(s.cfg.HttpStatus.Warning)
		w.Header().Set("Content-Type", "application/text")

		for idx, warn := range warns {
			line := fmt.Sprintf("%d: %s\n", idx, warn)
			_, err := w.Write([]byte(line))
			if err != nil {
				l.Error("Failed to write the response body",
					zap.Error(err),
				)
			}
		}

		l.Warn("Healthcheck encountered upstream error(s)",
			zap.Error(errors.Join(errs...)),
			zap.Int("http_status", s.cfg.HttpStatus.Warning),
		)
	}
}
