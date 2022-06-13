package main

import (
	"flag"
	"fmt"
	"github.com/pralolik/templgrid/cmd"
	"github.com/pralolik/templgrid/cmd/generator/input"
	"github.com/pralolik/templgrid/cmd/generator/output"
	"github.com/pralolik/templgrid/cmd/generator/output/auth"
	"os"
	"time"

	"github.com/pralolik/templgrid/cmd/generator"
	"github.com/pralolik/templgrid/cmd/logging"
)

// Variables which are related to Version command.
// Should be specified by '-ldflags' during the build phase.
// Example:
// GOOS=linux GOARCH=amd64 go build -ldflags="-X main.Version=$VERSION \
var (
	Version   = "unknown"
	BuildTime = time.Now().Format(time.RFC822)
)

func main() {
	flag.Parse()
	cfg, err := cmd.NewConfig(os.Args)
	if err != nil {
		fmt.Printf("Error with cfg: '%v'\n", err)
		os.Exit(1)
	}

	var log logging.Logger
	if cfg.Command != cmd.VersionCmd {
		lvl, lvlErr := logging.ParseLevel(cfg.LogLvl)
		if lvlErr != nil {
			fmt.Printf("Error with cfg: '%v'\n", lvlErr)
			os.Exit(1)
		}

		log = logging.NewStdLog(lvl)
	}
	log.Debug("Config %v", cfg)
	switch cfg.Command {
	case cmd.VersionCmd:
		versionCommand()
		os.Exit(0)
	case cmd.GenerateCmd:
		if err := generateCommand(log, cfg); err != nil {
			log.Error("Generate command error: %v", err)
			os.Exit(1)
		}
		os.Exit(0)
	default:
		log.Error("Following command expected: '%v'", cmd.AvailableCommands)
		os.Exit(1)
	}
}

func generateCommand(log logging.Logger, cfg *cmd.Config) error {
	genInpt, err := newInput(log, cfg)
	if err != nil {
		return fmt.Errorf("generator creation error: %v", err)
	}
	log.Debug("Generation input created")
	genOtpts, err := newOutputs(log, cfg)
	if err != nil {
		return fmt.Errorf("generator creation error: %v", err)
	}
	log.Debug("Generation output created")
	tmlGenerator := generator.New(genInpt, genOtpts, log)
	if err := tmlGenerator.Generate(); err != nil {
		return fmt.Errorf("generation error: %v", err)
	}

	return nil
}

func versionCommand() {
	fmt.Printf("Version: %s\n", Version)
	fmt.Printf("Built on: %s\n", BuildTime)
}

func newInput(log logging.Logger, cfg *cmd.Config) (input.Interface, error) {
	var inpt input.Interface

	if cfg.FileInput.Enabled {
		inpt = input.NewDirectoryInput(cfg.Prefix, cfg.FileInput.EmailPath, cfg.FileInput.ComponentPath, log)
	}

	if inpt == nil {
		return nil, fmt.Errorf("no input source for generator")
	}

	return inpt, nil
}

func newOutputs(log logging.Logger, cfg *cmd.Config) ([]output.Interface, error) {
	var otpts []output.Interface

	if cfg.FileOutput.Enabled {
		otpts = append(otpts, output.NewDirectoryOutput(cfg.FileOutput.Path, log))
	}
	if cfg.PreviewOutput.Enabled {
		otpts = append(otpts, output.NewPreviewOutput(cfg.PreviewOutput.Path, log))
	}

	if cfg.ApiSendgridOutput.Enabled {
		apiSndCfg := cfg.ApiSendgridOutput
		apiOutput, err := output.NewApiOutput(
			getApiAuthProvider(apiSndCfg.ApiConfig.Auth),
			apiSndCfg.DeleteNotListed,
			apiSndCfg.ApiConfig.Schema,
			apiSndCfg.ApiConfig.Host,
			apiSndCfg.ApiConfig.GetPath,
			apiSndCfg.ApiConfig.PostPath,
			apiSndCfg.ApiConfig.DeletePath,
			log,
		)
		if err != nil {
			return nil, fmt.Errorf("api-sendgrid output creation error %v", err)
		}

		if apiSndCfg.SendgridConfig.Enabled {
			otpts = append(
				otpts,
				output.NewSendGridOutput(
					apiOutput,
					apiSndCfg.SendgridConfig.Host,
					apiSndCfg.SendgridConfig.ApiKey,
					apiSndCfg.SendgridConfig.ActivateNewVersion,
				),
			)
		} else {
			otpts = append(otpts, apiOutput)
		}
	}

	if len(otpts) == 0 {
		return nil, fmt.Errorf("no output source for generator")
	}

	return otpts, nil
}

func getApiAuthProvider(_ cmd.AuthConfig) auth.Provider {
	return auth.NoneAuth{}
}
