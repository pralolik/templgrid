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
	Command           string
	LogLvl            string            `yaml:"logLvl"`
	Prefix            string            `yaml:"prefix"`
	FileInput         fileInput         `yaml:"file-input"`
	FileOutput        fileOutput        `yaml:"file-output"`
	PreviewOutput     fileOutput        `yaml:"preview-output"`
	ApiSendgridOutput apiSendgridOutput `yaml:"api-sendgrid-output"`
}

func (c *Config) validate() error {
	// todo config validation
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

type apiSendgridOutput struct {
	Enabled         bool           `yaml:"enabled"`
	DeleteNotListed bool           `yaml:"delete-not-listed"`
	ApiConfig       apiConfig      `yaml:"api"`
	SendgridConfig  sendgridConfig `yaml:"sendgrid"`
}

type sendgridConfig struct {
	Enabled            bool   `yaml:"enabled"`
	Host               string `yaml:"host"`
	ApiKey             string `yaml:"api-key"`
	ActivateNewVersion bool   `yaml:"activate-new-version"`
}

type apiConfig struct {
	Auth       AuthConfig `yaml:"auth"`
	Schema     string     `yaml:"schema"`
	Host       string     `yaml:"host"`
	GetPath    string     `yaml:"get-path"`
	PostPath   string     `yaml:"post-path"`
	DeletePath string     `yaml:"delete-path"`
}

type authType string

func (at *authType) IsInEnum() bool {
	switch *at {
	case BearerAuthType, BasicAuthType, HmacAuthType, NoneAuthType:
		return true
	default:
		return false
	}
}

const (
	BearerAuthType authType = "bearer"
	BasicAuthType  authType = "basic"
	HmacAuthType   authType = "hmac"
	NoneAuthType   authType = "none"
)

type AuthConfig struct {
	AuthType authType       `yaml:"auth-type"`
	Hmac     hmacAuthType   `yaml:"hmac"`
	Basic    basicAuthType  `yaml:"basic"`
	Bearer   bearerAuthType `yaml:"bearer"`
}

type bearerAuthType struct {
	Token  string `yaml:"token"`
	Header string `yaml:"header-name"`
}

type basicAuthType struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type hmacAuthType struct {
	key string `yaml:"key"`
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
