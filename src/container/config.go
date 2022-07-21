package container

import (
	"flag"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	VersionCmd = "version"
	RunCommand = "run"
)

var AvailableCommands = []string{VersionCmd, RunCommand}

type Config struct {
	Command   string
	LogLvl    string         `yaml:"logLvl"`
	APIConfig apiConfig      `yaml:"api"`
	Sendgrid  sendgridConfig `yaml:"sendgrid"`
}

func (c *Config) validate() error {
	return nil
}

type apiConfig struct {
	APIKey  string `yaml:"api-key"`
	Port    string `yaml:"port"`
	Enabled bool   `yaml:"enabled"`
}

type sendgridConfig struct {
	PrivateToken string `yaml:"private-token"`
	Enabled      bool   `yaml:"enabled"`
	SandBox      bool   `yaml:"sand-box"`
}

func NewConfig(args []string) (*Config, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("following command expected: '%v' ", AvailableCommands)
	}
	cfg := &Config{
		Command: os.Args[1],
	}

	if cfg.Command == VersionCmd {
		return cfg, nil
	}

	cmd := flag.NewFlagSet(cfg.Command, flag.ExitOnError)
	var configPath string
	cmd.StringVar(&configPath, "f", "", "configs path")
	if err := cmd.Parse(args[2:]); err != nil {
		return nil, err
	}

	if configPath != "" {
		v, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("can't read file: %w ", err)
		}
		if err := yaml.Unmarshal(v, &cfg); err != nil {
			return nil, fmt.Errorf("can't unmarshal file: %w ", err)
		}
	}

	// todo add rewrite of configs via console
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters: %w ", err)
	}

	return cfg, nil
}
