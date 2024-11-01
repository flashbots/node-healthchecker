package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/flashbots/node-healthchecker/healthcheck"
	"github.com/flashbots/node-healthchecker/logutils"
	"go.uber.org/zap"
)

func (s *Server) healthcheck(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Healthcheck.CacheTimeout != 0 {
		s.cache.mx.Lock()
		defer s.cache.mx.Unlock()

		now := time.Now()

		if s.cache.expiry.After(now) {
			s.report(w, r, s.cache.errs, s.cache.wrns)
			return
		}

		s.cache.expiry = now.Add(s.cfg.Healthcheck.CacheTimeout)
	}

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
	wrns := []error{}
	for count > 0 {
		count--
		if res := <-results; res != nil {
			if !res.Ok {
				errs = append(errs, res.Err)
			} else if res.Err != nil {
				wrns = append(wrns, res.Err)
			}
		}
	}
	close(results)

	s.report(w, r, errs, wrns)

	if s.cfg.Healthcheck.CacheTimeout != 0 {
		s.cache.errs = errs
		s.cache.wrns = wrns
	}
}

func (s *Server) report(w http.ResponseWriter, r *http.Request, errs, wrns []error) {
	l := logutils.LoggerFromRequest(r)

	switch {
	case len(errs) == 0 && len(wrns) == 0:
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
		for idx, warn := range wrns {
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

	case len(errs) == 0 && len(wrns) > 0:
		w.WriteHeader(s.cfg.HttpStatus.Warning)
		w.Header().Set("Content-Type", "application/text")

		for idx, warn := range wrns {
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
