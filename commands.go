package main

import (
	"fmt"
	"os"
	"os/signal"

	appGo "github.com/kuuyee/otto-learn/builtin/app/go"
	appRuby "github.com/kuuyee/otto-learn/builtin/app/ruby"
	foundationConsul "github.com/kuuyee/otto-learn/builtin/foundation/consul"
	infraAws "github.com/kuuyee/otto-learn/builtin/infra/aws"

	"github.com/hashicorp/otto/appfile/detect"
	"github.com/kuuyee/otto-learn/app"
	"github.com/kuuyee/otto-learn/command"
	"github.com/kuuyee/otto-learn/foundation"
	"github.com/kuuyee/otto-learn/infrastructure"
	"github.com/kuuyee/otto-learn/otto"
	"github.com/mitchellh/cli"
)

// Commands map所有的Otto命令
var Commands map[string]cli.CommandFactory
var CommandsInclude []string

// 定义otto识别的开发语言类型
var Detectors = []*detect.Detector{
	&detect.Detector{
		Type: "go",
		File: []string{"*.go"},
	},
	&detect.Detector{
		Type: "php",
		File: []string{"*.php", "composer.json"},
	},
	&detect.Detector{
		Type: "rails",
		File: []string{"config/application.rb"},
	},
	&detect.Detector{
		Type: "ruby",
		File: []string{"*.rb", "Gemfile", "config.ru"},
	},
	&detect.Detector{
		Type: "node",
		File: []string{"package.json"},
	},
}

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

	apps := appGo.Tuples.Map(app.StructFactory(new(appGo.App)))
	apps.Add(appRuby.Tuples.Map(app.StructFactory(new(appRuby.App))))

	foundations := foundationConsul.Tuples.Map(foundation.StructFactory(new(foundationConsul.Foundation)))
	fmt.Println(foundations)

	meta := command.Meta{
		CoreConfig: &otto.CoreConfig{
			Apps:        apps,
			Foundations: foundations,
			Infrastructures: map[string]infrastructure.Factory{
				"aws": infraAws.Infra,
			},
		},
		Ui: Ui,
	}
	fmt.Println(meta)

	CommandsInclude = []string{
		"compile",
		"build",
		"deploy",
		"dev",
		"infra",
		"status",
		"version",
	}

	Commands = map[string]cli.CommandFactory{
		"compile": func() (cli.Command, error) {
			return &command.CompileCommand{
				Meta:      meta,
				Detectors: Detectors,
			}, nil
		},
	}
}

// makeShutdownCh创建一个终端监听并返回一个channel。
// 当监听到终端时发送一条消息给channel.
func makeShutdownCh() <-chan struct{} {
	resultCh := make(chan struct{})

	signalCh := make(chan os.Signal, 4)
	signal.Notify(signalCh, os.Interrupt)
	go func() {
		for {
			<-signalCh
			resultCh <- struct{}{}
		}
	}()

	return resultCh
}
