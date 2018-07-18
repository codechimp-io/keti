package config

import (
	"fmt"

	"github.com/codechimp-io/keti/log"

	"github.com/kelseyhightower/envconfig"
)

func init() {
	err := envconfig.Process("", &Options)
	if err != nil {
		log.Fatalf("Error loading config: %s", err)
	}
}

type EnvConfig struct {
	Debug bool   `envconfig:"KETI_DEBUG" default:"false" required:"true"`
	Token string `envconfig:"KETI_BOT_TOKEN" default:""`
}

func (e *EnvConfig) BotToken() string {
	if e.Token == "" {
		log.Fatal("Discord Bot token cannot be left blank")
	}

	return fmt.Sprintf("Bot %s", e.Token)
}

var Options EnvConfig
