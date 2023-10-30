package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	version = "development"
)

func main() {
	var logFormat, logLevel string

	app := &cli.App{
		Name:    "node-healthchecker",
		Usage:   "Aggregates complex health-checks of a blockchain node into a simple http endpoint",
		Version: version,

		Action: func(c *cli.Context) error {
			return cli.ShowAppHelp(c)
		},

		Flags: []cli.Flag{
			&cli.StringFlag{
				Destination: &logLevel,
				EnvVars:     []string{"LOG_LEVEL"},
				Name:        "log-level",
				Usage:       "logging level",
				Value:       "info",
			},

			&cli.StringFlag{
				Destination: &logFormat,
				EnvVars:     []string{"LOG_MODE"},
				Name:        "log-mode",
				Usage:       "logging mode",
				Value:       "prod",
			},
		},

		Before: func(ctx *cli.Context) error {
			err := setupLogger(logLevel, logFormat)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to configure the logging: %s\n", err)
			}
			return err
		},

		Commands: []*cli.Command{
			CommandServe(),
		},
	}

	defer func() {
		zap.L().Sync()
	}()
	if err := app.Run(os.Args); err != nil {
		zap.L().Error("Failed with error", zap.Error(err))
	}
}

func setupLogger(level, mode string) error {
	var config zap.Config
	switch strings.ToLower(mode) {
	case "dev":
		config = zap.NewDevelopmentConfig()
	case "prod":
		config = zap.NewProductionConfig()
	default:
		return fmt.Errorf("invalid log-mode '%s'", mode)
	}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log-level '%s': %w", level, err)
	}
	config.Level = logLevel

	l, err := config.Build()
	if err != nil {
		return fmt.Errorf("failed to build the logger: %w", err)
	}
	zap.ReplaceGlobals(l)

	return nil
}
