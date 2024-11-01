package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/flashbots/node-healthchecker/config"
	"github.com/flashbots/node-healthchecker/healthcheck"
	"github.com/flashbots/node-healthchecker/httplogger"
	"github.com/flashbots/node-healthchecker/logutils"
)

type Server struct {
	cfg *config.Config

	failure chan error

	logger *zap.Logger
	server *http.Server

	cache    *cache
	monitors []healthcheck.Monitor
}

func New(cfg *config.Config) (*Server, error) {
	monitors := make([]healthcheck.Monitor, 0)

	if cfg.HealthcheckGeth.BaseURL != "" {
		monitors = append(monitors, func(ctx context.Context) *healthcheck.Result {
			return healthcheck.Geth(ctx, &cfg.HealthcheckGeth)
		})
	}

	if cfg.HealthcheckLighthouse.BaseURL != "" {
		monitors = append(monitors, func(ctx context.Context) *healthcheck.Result {
			return healthcheck.Lighthouse(ctx, &cfg.HealthcheckLighthouse)
		})
	}

	if cfg.HealthcheckOpNode.BaseURL != "" {
		monitors = append(monitors, func(ctx context.Context) *healthcheck.Result {
			return healthcheck.OpNode(ctx, &cfg.HealthcheckOpNode)
		})
	}

	if cfg.HealthcheckReth.BaseURL != "" {
		monitors = append(monitors, func(ctx context.Context) *healthcheck.Result {
			return healthcheck.Reth(ctx, &cfg.HealthcheckReth)
		})
	}

	s := &Server{
		cfg:      cfg,
		failure:  make(chan error, 1),
		logger:   zap.L(),
		monitors: monitors,
	}

	if cfg.Healthcheck.CacheCoolOff != 0 {
		s.cache = &cache{}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.healthcheck)
	mux.Handle("/metrics", promhttp.Handler())
	handler := httplogger.Middleware(s.logger, mux)

	s.server = &http.Server{
		Addr:              cfg.Server.ListenAddress,
		ErrorLog:          logutils.NewHttpServerErrorLogger(s.logger),
		Handler:           handler,
		MaxHeaderBytes:    1024,
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	return s, nil
}

func (s *Server) Run() error {
	l := s.logger
	ctx := logutils.ContextWithLogger(context.Background(), l)

	go func() { // run the server
		l.Info("Blockchain node healthchecker server is going up...",
			zap.String("server_listen_address", s.cfg.Server.ListenAddress),
		)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.failure <- err
		}
		l.Info("Blockchain node healthchecker server is down")
	}()

	errs := []error{}
	{ // wait until termination or internal failure
		terminator := make(chan os.Signal, 1)
		signal.Notify(terminator, os.Interrupt, syscall.SIGTERM)

		select {
		case stop := <-terminator:
			l.Info("Stop signal received; shutting down...",
				zap.String("signal", stop.String()),
			)
		case err := <-s.failure:
			l.Error("Internal failure; shutting down...",
				zap.Error(err),
			)
			errs = append(errs, err)
		exhaustErrors:
			for { // exhaust the errors
				select {
				case err := <-s.failure:
					l.Error("Extra internal failure",
						zap.Error(err),
					)
					errs = append(errs, err)
				default:
					break exhaustErrors
				}
			}
		}
	}

	{ // stop the server
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			l.Error("Blockchain node healthchecker server shutdown failed",
				zap.Error(err),
			)
		}
	}

	switch len(errs) {
	default:
		return errors.Join(errs...)
	case 1:
		return errs[0]
	case 0:
		return nil
	}
}
