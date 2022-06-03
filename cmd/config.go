package cmd

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

const (
	VersionCmd  = "version"
	GenerateCmd = "generate"
)

var AvailableCommands = []string{VersionCmd, GenerateCmd}

type Config struct {
	Command        string
	LogLvl         string         `yaml:"logLvl"`
	FileInput      fileInput      `yaml:"file-input"`
	FileOutput     fileOutput     `yaml:"file-output"`
	PreviewOutput  fileOutput     `yaml:"preview-output"`
	SendgridOutput sendgridOutput `yaml:"sendgrid-output"`
}

func (c *Config) validate() error {
	return nil
}

type fileInput struct {
	Enabled       bool   `yaml:"enabled"`
	EmailPath     string `yaml:"email-path"`
	ComponentPath string `yaml:"component-path"`
}

type fileOutput struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"result-path"`
}

type sendgridOutput struct {
	Enabled        bool           `yaml:"enabled"`
	SendgridConfig sendgridConfig `yaml:"sendgrid"`
	ApiConfig      apiConfig      `yaml:"api"`
}

type sendgridConfig struct {
	Enabled      bool   `yaml:"enabled"`
	PrivateToken string `yaml:"private-token"`
}

type apiConfig struct {
	Auth     authConfig `yaml:"auth"`
	Schema   string     `yaml:"schema"`
	Host     string     `yaml:"host"`
	GetPath  string     `yaml:"get-path"`
	SavePath string     `yaml:"save-path"`
}

type authType string

type authConfig struct {
	AuthType authType `yaml:"auth-type"`
	Token    string   `yaml:"secret"`
}

func NewConfig(args []string) (*Config, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("following command expected: '%v'\n", AvailableCommands)
	}
	cfg := &Config{
		Command: os.Args[1],
	}

	if cfg.Command == VersionCmd {
		return cfg, nil
	}

	cmd := flag.NewFlagSet(cfg.Command, flag.ExitOnError)
	var configPath string
	cmd.StringVar(&configPath, "f", "", "config path")
	if err := cmd.Parse(args[2:]); err != nil {
		return nil, err
	}

	if configPath != "" {
		v, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("can't read file %v", err)
		}
		if err := yaml.Unmarshal(v, &cfg); err != nil {
			return nil, fmt.Errorf("can't unmarshal file %v", err)
		}
	}

	// todo add rewrite of config via console
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid parameters %v", err)
	}

	return cfg, nil
}
