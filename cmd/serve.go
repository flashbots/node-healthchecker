package main

import (
	"time"

	"github.com/flashbots/node-healthchecker/healthchecker"
	"github.com/urfave/cli/v2"
)

func CommandServe() *cli.Command {
	var cfg healthchecker.Config

	return &cli.Command{
		Name:  "serve",
		Usage: "run the healthcheck server",

		Flags: []cli.Flag{

			// Serving

			&cli.StringFlag{
				Category:    "Serving:",
				Destination: &cfg.ServeAddress,
				Name:        "serve-address",
				Usage:       "respond to health-checks at the address of `host:port`",
				Value:       "0.0.0.0:8080",
			},

			&cli.DurationFlag{
				Category:    "Serving:",
				Destination: &cfg.Timeout,
				Name:        "timeout",
				Usage:       "healthcheck(s) timeout `duration`",
				Value:       time.Second,
			},

			// Monitoring

			&cli.StringFlag{
				Category:    "Monitoring:",
				Destination: &cfg.MonitorGethURL,
				Name:        "monitor-geth-url",
				Usage:       "monitor geth sync-status via RPC at specified `URL`",
			},

			&cli.StringFlag{
				Category:    "Monitoring:",
				Destination: &cfg.MonitorLighthouseURL,
				Name:        "monitor-lighthouse-url",
				Usage:       "monitor lighthouse sync-status via RPC at specified `URL`",
			},
		},

		Action: func(ctx *cli.Context) error {
			h, err := healthchecker.New(&cfg)
			if err != nil {
				return err
			}
			return h.Serve()
		},
	}
}
