package main

import (
	"github.com/urfave/cli/v2"

	"github.com/flashbots/node-healthchecker/config"
)

func CommandHelp(_ *config.Config) *cli.Command {
	return &cli.Command{
		Usage: "show the list of commands or help for one command",
		Name:  "help",

		Action: func(clictx *cli.Context) error {
			return cli.ShowAppHelp(clictx)
		},
	}
}
