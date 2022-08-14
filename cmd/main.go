package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pralolik/templgrid/src/container"
	"github.com/pralolik/templgrid/src/logging"
)

// Version Variables which are related to Version command.
// Should be specified by '-ldflags' during the build phase.
// Example:
// GOOS=linux GOARCH=amd64 go build -ldflags="-X main.Version=$VERSION" -o templgrid.
var Version = "unknown"

func main() {
	flag.Parse()
	cfg, err := container.NewConfig(os.Args)
	if err != nil {
		fmt.Printf("Error with cfg: '%v'\n", err)
		os.Exit(1)
	}

	var log logging.Logger
	if cfg.Command != container.VersionCmd {
		lvl, lvlErr := logging.ParseLevel(cfg.LogLvl)
		if lvlErr != nil {
			log.Error("Error with lvlLog (Info by default): '%v'\n", lvlErr)
		}

		log = logging.NewStdLog(lvl)
		log.Debug("Config: %v ", cfg)
	}
	switch cfg.Command {
	case container.VersionCmd:
		versionCommand()
		os.Exit(0)
	case container.RunCommand:
		if err := runCommand(log, cfg); err != nil {
			log.Error("Run command error: %v ", err)
			os.Exit(1)
		}
		os.Exit(0)
	default:
		log.Error("Following command expected: '%v'", container.AvailableCommands)
		os.Exit(1)
	}
}

func runCommand(log logging.Logger, cfg *container.Config) error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	appContainer, err := container.NewAppContainer(cfg, log)
	if err != nil {
		return fmt.Errorf("app container creation error: %w ", err)
	}
	appContainer.Run(ctx)

	return nil
}

func versionCommand() {
	fmt.Printf("Version: %s\n", Version)
}
