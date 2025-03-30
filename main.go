package main

import (
	"fmt"
	"os"

	kj "github.com/nr8-io/argocd-konn-jsonnet-plugin/internal/plugin/konn-jsonnet"
	"github.com/nr8-io/argocd-konn-jsonnet-plugin/pkg/zino"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: init | generate")
		os.Exit(1)
	}

	// get log level from env
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "debug"
	}

	log, err := zino.NewLogger(level)
	if err != nil {
		panic(err)
	}

	// create new konn jsonnet plugin with logger
	plugin, err := kj.NewKonnJsonnetPlugin(
		kj.WithLogger(log),
	)

	if err != nil {
		log.Error().Err(err).Msg("Failed to create plugin")
		os.Exit(1)
	}

	// init plugin
	if os.Args[1] == "init" {
		log.Debug().Interface("plugin", plugin).Msg("Init")

		err := plugin.Init()
		if err != nil {
			log.Error().Err(err).Msg("Failed to init plugin")
			os.Exit(1)
		}
		return
	}

	// generate argocd manifest
	if os.Args[1] == "generate" {
		err := plugin.Generate()
		if err != nil {
			os.Exit(1)
		}
		return
	}
}
