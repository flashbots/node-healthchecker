package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"time"

	"github.com/flashbots/node-healthchecker/healthcheck"
	"github.com/flashbots/node-healthchecker/logutils"
	"github.com/flashbots/node-healthchecker/metrics"

	"go.opentelemetry.io/otel/attribute"
	otelapi "go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

func (s *Server) healthcheck(w http.ResponseWriter, r *http.Request) {
	if s.cfg.Healthcheck.CacheCoolOff != 0 {
		s.cache.mx.Lock()
		defer s.cache.mx.Unlock()

		now := time.Now()

		if s.cache.expiry.After(now) {
			s.report(w, r, true, s.cache.errs, s.cache.wrns)
			return
		}

		s.cache.expiry = now.Add(s.cfg.Healthcheck.CacheCoolOff)
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
				errs = append(errs, res.Error())

				metrics.HealthcheckUp.Record(context.Background(), 0, otelapi.WithAttributes(
					attribute.KeyValue{Key: "healthcheck_source", Value: attribute.StringValue(res.Source)},
				))

				metrics.HealthchecksNokCount.Add(context.Background(), 1, otelapi.WithAttributes(
					attribute.KeyValue{Key: "healthcheck_source", Value: attribute.StringValue(res.Source)},
				))

				if s.ok[res.Source] {
					s.ok[res.Source] = false
					metrics.HealthchecksFlipCount.Add(context.Background(), 1, otelapi.WithAttributes(
						attribute.KeyValue{Key: "healthcheck_source", Value: attribute.StringValue(res.Source)},
					))
				}

				continue
			}

			if res.Err != nil {
				wrns = append(wrns, res.Error())
			}

			metrics.HealthcheckUp.Record(context.Background(), 1, otelapi.WithAttributes(
				attribute.KeyValue{Key: "healthcheck_source", Value: attribute.StringValue(res.Source)},
			))

			metrics.HealthchecksOkCount.Add(context.Background(), 1, otelapi.WithAttributes(
				attribute.KeyValue{Key: "healthcheck_source", Value: attribute.StringValue(res.Source)},
			))

			if !s.ok[res.Source] {
				s.ok[res.Source] = true
				metrics.HealthchecksFlipCount.Add(context.Background(), 1, otelapi.WithAttributes(
					attribute.KeyValue{Key: "healthcheck_source", Value: attribute.StringValue(res.Source)},
				))
			}
		}
	}
	close(results)

	s.report(w, r, false, errs, wrns)

	if s.cfg.Healthcheck.CacheCoolOff != 0 {
		s.cache.errs = errs
		s.cache.wrns = append(wrns, errors.New("cached healthcheck"))
	}
}

func (s *Server) report(w http.ResponseWriter, r *http.Request, cached bool, errs, wrns []error) {
	l := logutils.LoggerFromRequest(r)

	if cached {
		l.Debug("Sending cached healthcheck")
	}

	switch {
	case len(errs) == 0 && len(wrns) == 0:
		w.WriteHeader(s.cfg.HttpStatus.Ok)
		return

	case len(errs) > 0:
		w.WriteHeader(s.cfg.HttpStatus.Error)
		w.Header().Set("Content-Type", "application/text")

		for idx, err := range errs {
			line := fmt.Sprintf("%d: error: %s\n", idx, err)
			if _, _err := w.Write([]byte(line)); _err != nil {
				l.Error("Failed to write the response body",
					zap.Error(_err),
				)
			}
		}
		offset := len(errs)
		for idx, warn := range wrns {
			line := fmt.Sprintf("%d: warning: %s\n", offset+idx, warn)
			if _, _err := w.Write([]byte(line)); _err != nil {
				l.Error("Failed to write the response body",
					zap.Error(_err),
				)
			}
		}

		if !cached {
			l.Warn("Healthcheck encountered upstream error(s)",
				zap.Error(errors.Join(errs...)),
				zap.Int("http_status", s.cfg.HttpStatus.Error),
			)
		}

	case len(errs) == 0 && len(wrns) > 0:
		w.WriteHeader(s.cfg.HttpStatus.Warning)
		w.Header().Set("Content-Type", "application/text")

		for idx, warn := range wrns {
			line := fmt.Sprintf("%d: warning: %s\n", idx, warn)
			if _, _err := w.Write([]byte(line)); _err != nil {
				l.Error("Failed to write the response body",
					zap.Error(_err),
				)
			}
		}

		if !cached {
			l.Warn("Healthcheck encountered upstream warning(s)",
				zap.Error(errors.Join(wrns...)),
				zap.Int("http_status", s.cfg.HttpStatus.Warning),
			)
		}
	}
}
