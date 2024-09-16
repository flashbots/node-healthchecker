package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/flashbots/node-healthchecker/config"
	"github.com/flashbots/node-healthchecker/logutils"
)

var (
	version = "development"
)

const (
	envPrefix = "NH_"
)

func main() {
	cfg := &config.Config{}

	flags := []cli.Flag{
		&cli.StringFlag{
			Destination: &cfg.Log.Level,
			EnvVars:     []string{envPrefix + "LOG_LEVEL"},
			Name:        "log-level",
			Usage:       "logging level",
			Value:       "info",
		},

		&cli.StringFlag{
			Destination: &cfg.Log.Mode,
			EnvVars:     []string{envPrefix + "LOG_MODE"},
			Name:        "log-mode",
			Usage:       "logging mode",
			Value:       "prod",
		},
	}

	commands := []*cli.Command{
		CommandServe(cfg),
		CommandHelp(cfg),
	}

	app := &cli.App{
		Name:    "node-healthchecker",
		Usage:   "Report the sync-status of a blockchain node as HTTP status",
		Version: version,

		Flags:          flags,
		Commands:       commands,
		DefaultCommand: commands[0].Name,

		Before: func(_ *cli.Context) error {
			// setup logger
			l, err := logutils.NewLogger(&cfg.Log)
			if err != nil {
				return err
			}
			zap.ReplaceGlobals(l)

			return nil
		},

		Action: func(clictx *cli.Context) error {
			return cli.ShowAppHelp(clictx)
		},
	}

	defer func() {
		zap.L().Sync() //nolint:errcheck
	}()
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "\nFailed with error:\n\n%s\n\n", err.Error())
		os.Exit(1)
	}
}
