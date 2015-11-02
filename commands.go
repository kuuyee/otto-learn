package main

import (
	"github.com/mitchellh/cli"
	"os"
)

// Commands map所有的Otto命令
var Commands map[string]cli.CommandFactory
var CommandsInclude []string

// Ui是cli.Ui,用来和外界交互
var Ui cli.Ui

const (
	ErrorPrefix  = "e:"
	OutputPrefix = "o:"
)

func init() {
	Ui = &cli.ColoredUi{
		OutputColor: cli.UiColorNone,
		InfoColor:   cli.UiColorNone,
		ErrorColor:  cli.UiColorRed,
		WarnColor:   cli.UiColorNone,
		Ui: &cli.PrefixedUi{
			AskPrefix:    OutputPrefix,
			OutputPrefix: OutputPrefix,
			InfoPrefix:   OutputPrefix,
			ErrorPrefix:  ErrorPrefix,
			Ui:           &cli.BasicUi{Writer: os.Stdout},
		},
	}

	//apps:=appGo.

	CommandsInclude = []string{
		"compile",
		"build",
		"deploy",
		"dev",
		"infra",
		"status",
		"version",
	}
}
