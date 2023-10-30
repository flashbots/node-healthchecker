package healthchecker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flashbots/node-healthchecker/httplogger"
	"go.uber.org/zap"
)

type Healthchecker struct {
	addr    string
	log     *zap.Logger
	timeout time.Duration

	monitors []monitor
}

type Config struct {
	MonitorGethURL       string
	MonitorLighthouseURL string
	ServeAddress         string
	Timeout              time.Duration
}

type monitor = func(context.Context) error

func New(cfg *Config) (*Healthchecker, error) {
	h := &Healthchecker{
		addr:    cfg.ServeAddress,
		log:     zap.L(),
		timeout: cfg.Timeout,
	}

	// Configure geth checks

	if cfg.MonitorGethURL != "" {
		rpcURL, err := url.JoinPath(cfg.MonitorGethURL, "/")
		if err != nil {
			return nil, err
		}
		h.monitors = append(h.monitors, func(ctx context.Context) error {
			if err := h.checkGeth(ctx, rpcURL); err != nil {
				return fmt.Errorf(
					"error while checking sync-status of geth at '%s': %w",
					rpcURL,
					err,
				)
			}
			return nil
		})
	}

	// Configure lighthouse checks

	if cfg.MonitorLighthouseURL != "" {
		syncingURL, err := url.JoinPath(cfg.MonitorLighthouseURL, "lighthouse/syncing")
		if err != nil {
			return nil, err
		}
		h.monitors = append(h.monitors, func(ctx context.Context) error {
			if err := h.checkLighthouse(ctx, syncingURL); err != nil {
				return fmt.Errorf(
					"error while checking sync-status of lighthouse at '%s': %w",
					syncingURL,
					err,
				)
			}
			return nil
		})
	}

	return h, nil
}

func (h *Healthchecker) Serve() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", h.handleHTTPRequest)
	handler := httplogger.Middleware(h.log, mux)

	srv := &http.Server{
		Addr:              h.addr,
		Handler:           handler,
		MaxHeaderBytes:    1024,
		ReadHeaderTimeout: 30 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
	}

	go func() {
		terminator := make(chan os.Signal, 1)
		signal.Notify(terminator, os.Interrupt, syscall.SIGTERM)
		s := <-terminator

		h.log.Info("Stop signal received; shutting down...", zap.String("signal", s.String()))
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			h.log.Error("HTTP server shutdown failed", zap.Error(err))
		}
	}()

	h.log.Info("Starting up...", zap.String("address", h.addr))
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		h.log.Error("HTTP server failed", zap.Error(err))
	}
	h.log.Info("Server is down")

	return nil
}

func (h *Healthchecker) handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	count := len(h.monitors)
	results := make(chan error, count)

	for _, m := range h.monitors {
		monitor := m // https://go.dev/blog/loopvar-preview
		ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
		defer cancel()
		go func() {
			results <- monitor(ctx)
		}()
	}

	errors := []error{}
	for count > 0 {
		count--
		if res := <-results; res != nil {
			errors = append(errors, res)
		}
	}
	close(results)

	if len(errors) == 0 {
		return
	}

	w.Header().Set("Content-Type", "application/text")
	w.WriteHeader(http.StatusInternalServerError)
	for idx, err := range errors {
		line := fmt.Sprintf("%d: %s\n", idx, err)
		w.Write([]byte(line))
	}
}
