package main

import (
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/flashbots/node-healthchecker/config"
	"github.com/flashbots/node-healthchecker/server"
	"github.com/flashbots/node-healthchecker/utils"
)

const (
	categoryHealthcheck           = "healthcheck"
	categoryHealthcheckGeth       = "healthcheck geth"
	categoryHealthcheckLighthouse = "healthcheck lighthouse"
	categoryHealthcheckOpNode     = "healthcheck op-node"
	categoryHealthcheckReth       = "healthcheck reth"
	categoryHttpStatus            = "http status"
	categoryServer                = "server"
)

func CommandServe(cfg *config.Config) *cli.Command {
	ip := "0.0.0.0"
	if ipv4, err := utils.PrivateIPv4(); err == nil {
		ip = ipv4.String()
	}
	// healthcheck

	healthcheckFlags := []cli.Flag{ // --healthcheck-xxx
		&cli.DurationFlag{ // --healthcheck-block-age-threshold
			Category:    strings.ToUpper(categoryHealthcheck),
			Destination: &cfg.Healthcheck.BlockAgeThreshold,
			DefaultText: "disabled",
			EnvVars:     []string{envPrefix + strings.ToUpper(categoryHealthcheck) + "_BLOCK_AGE_THRESHOLD"},
			Name:        categoryHealthcheck + "-block-age-threshold",
			Usage:       "monitor the age of latest block and report unhealthy if it's over specified `duration`",
			Value:       0,
		},

		&cli.DurationFlag{ // --healthcheck-cache-cool-off
			Category:    strings.ToUpper(categoryHealthcheck),
			Destination: &cfg.Healthcheck.CacheCoolOff,
			EnvVars:     []string{envPrefix + strings.ToUpper(categoryHealthcheck) + "_CACHE_COOL_OFF"},
			Name:        categoryHealthcheck + "-cache-cool-off",
			Usage:       "re-use healthcheck results for the specified `duration`",
			Value:       750 * time.Millisecond,
		},

		&cli.DurationFlag{ // --healthcheck-timeout
			Category:    strings.ToUpper(categoryHealthcheck),
			Destination: &cfg.Healthcheck.Timeout,
			EnvVars:     []string{envPrefix + strings.ToUpper(categoryHealthcheck) + "_TIMEOUT"},
			Name:        categoryHealthcheck + "-timeout",
			Usage:       "maximum `duration` of a single healthcheck",
			Value:       time.Second,
		},

		&cli.DurationFlag{ // --healthcheck-unconditional-fail-duration
			Category:    strings.ToUpper(categoryHealthcheck),
			Destination: &cfg.Healthcheck.UnconditionalFailDuration,
			EnvVars:     []string{envPrefix + strings.ToUpper(categoryHealthcheck) + "_UNCONDITIONAL_FAIL_DURATION"},
			Name:        categoryHealthcheck + "-unconditional-fail-duration",
			DefaultText: "infinite",
			Usage:       "`duration` of node-healthchecker failing the healthchecks when requested with SIGHUP",
			Value:       0,
		},
	}

	// healthcheck geth

	healthcheckGethFlags := []cli.Flag{ // --healthcheck-geth-xxx
		&cli.StringFlag{ // --healthcheck-geth-base-url
			Category:    strings.ToUpper(categoryHealthcheckGeth),
			Destination: &cfg.HealthcheckGeth.BaseURL,
			EnvVars:     []string{envPrefix + strings.ReplaceAll(strings.ToUpper(categoryHealthcheckGeth), " ", "_") + "_BASE_URL"},
			Name:        strings.ReplaceAll(categoryHealthcheckGeth, " ", "-") + "-base-url",
			Usage:       "base `url` of geth's HTTP-RPC endpoint",
		},
	}

	// healthcheck lighthouse

	healthcheckLighthouseFlags := []cli.Flag{ // --healthcheck-lighthouse-xxx
		&cli.StringFlag{ // --healthcheck-lighthouse-base-url
			Category:    strings.ToUpper(categoryHealthcheckLighthouse),
			Destination: &cfg.HealthcheckLighthouse.BaseURL,
			EnvVars:     []string{envPrefix + strings.ReplaceAll(strings.ToUpper(categoryHealthcheckLighthouse), " ", "_") + "_BASE_URL"},
			Name:        strings.ReplaceAll(categoryHealthcheckLighthouse, " ", "-") + "-base-url",
			Usage:       "base `url` of lighthouse's HTTP-API endpoint",
		},
	}

	// healthcheck op-node

	healthcheckOpNodeFlags := []cli.Flag{ // --healthcheck-op-node-xxx
		&cli.StringFlag{ // --healthcheck-op-node-base-url
			Category:    strings.ToUpper(categoryHealthcheckOpNode),
			Destination: &cfg.HealthcheckOpNode.BaseURL,
			EnvVars:     []string{envPrefix + strings.ReplaceAll(strings.ReplaceAll(strings.ToUpper(categoryHealthcheckOpNode), " ", "_"), "-", "_") + "_BASE_URL"},
			Name:        strings.ReplaceAll(categoryHealthcheckOpNode, " ", "-") + "-base-url",
			Usage:       "base `url` of op-node's RPC endpoint",
		},

		&cli.Uint64Flag{ // --healthcheck-op-node-conf-distance
			Category:    strings.ToUpper(categoryHealthcheckOpNode),
			Destination: &cfg.HealthcheckOpNode.ConfirmationDistance,
			EnvVars:     []string{envPrefix + strings.ReplaceAll(strings.ReplaceAll(strings.ToUpper(categoryHealthcheckOpNode), " ", "_"), "-", "_") + "_CONF_DISTANCE"},
			Name:        strings.ReplaceAll(categoryHealthcheckOpNode, " ", "-") + "-conf-distance",
			Usage:       "number of l1 blocks that verifier keeps distance from the l1 head before deriving l2 data from",
			Value:       0,
		},
	}

	// healthcheck reth

	healthcheckRethFlags := []cli.Flag{ // --healthcheck-reth-xxx
		&cli.StringFlag{ // --healthcheck-reth-base-url
			Category:    strings.ToUpper(categoryHealthcheckReth),
			Destination: &cfg.HealthcheckReth.BaseURL,
			EnvVars:     []string{envPrefix + strings.ReplaceAll(strings.ToUpper(categoryHealthcheckReth), " ", "_") + "_BASE_URL"},
			Name:        strings.ReplaceAll(categoryHealthcheckReth, " ", "-") + "-base-url",
			Usage:       "base `url` of reth's HTTP-RPC endpoint",
		},
	}

	// http status

	httpStatusFlags := []cli.Flag{ // --healthcheck-http-status
		&cli.IntFlag{ // --healthcheck-http-status-ok
			Category:    strings.ToUpper(categoryHttpStatus),
			Destination: &cfg.HttpStatus.Ok,
			EnvVars:     []string{envPrefix + strings.ReplaceAll(strings.ToUpper(categoryHttpStatus), " ", "_") + "_OK"},
			Name:        strings.ReplaceAll(categoryHttpStatus, " ", "-") + "-ok",
			Usage:       "http `status` to report on good healthchecks",
			Value:       http.StatusOK,
		},

		&cli.IntFlag{ // --healthcheck-http-status-warning
			Category:    strings.ToUpper(categoryHttpStatus),
			Destination: &cfg.HttpStatus.Warning,
			EnvVars:     []string{envPrefix + strings.ReplaceAll(strings.ToUpper(categoryHttpStatus), " ", "_") + "_WARNING"},
			Name:        strings.ReplaceAll(categoryHttpStatus, " ", "-") + "-warning",
			Usage:       "http `status` to report on healthchecks with warnings",
			Value:       http.StatusAccepted,
		},

		&cli.IntFlag{ // --healthcheck-http-status-error
			Category:    strings.ToUpper(categoryHttpStatus),
			Destination: &cfg.HttpStatus.Error,
			EnvVars:     []string{envPrefix + strings.ReplaceAll(strings.ToUpper(categoryHttpStatus), " ", "_") + "_ERROR"},
			Name:        strings.ReplaceAll(categoryHttpStatus, " ", "-") + "-error",
			Usage:       "http `status` to report on healthchecks with errors",
			Value:       http.StatusInternalServerError,
		},
	}

	// server

	serverFlags := []cli.Flag{ // --server-xxx
		&cli.StringFlag{ // --server-listen-address
			Category:    strings.ToUpper(categoryServer),
			Destination: &cfg.Server.ListenAddress,
			EnvVars:     []string{envPrefix + strings.ToUpper(categoryServer) + "_LISTEN_ADDRESS"},
			Name:        categoryServer + "-listen-address",
			Usage:       "`host:port` for the server to listen on",
			Value:       ip + ":8080",
		},
	}

	return &cli.Command{
		Name:  "serve",
		Usage: "run node-healthchecker server",

		Flags: slices.Concat(
			healthcheckFlags,
			healthcheckGethFlags,
			healthcheckLighthouseFlags,
			healthcheckOpNodeFlags,
			healthcheckRethFlags,
			httpStatusFlags,
			serverFlags,
		),

		Before: func(ctx *cli.Context) error {
			if err := cfg.Preprocess(); err != nil {
				return err
			}
			return nil
		},

		Action: func(_ *cli.Context) error {
			s, err := server.New(cfg)
			if err != nil {
				return err
			}
			return s.Run()
		},
	}
}
