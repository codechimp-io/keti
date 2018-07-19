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
	Debug   bool `envconfig:"KETI_DEBUG" default:"false" required:"true"`
	Discord discord
}

type discord struct {
	Token      string `envconfig:"KETI_DISCORD_TOKEN" default:""`
	ShardID    int    `envconfig:"KETI_DISCORD_SHARD_ID" default:"0"`
	ShardCount int    `envconfig:"KETI_DISCORD_SHARD_COUNT" default:"1"`
	StatusChan string `envconfig:"KETI_DISCORD_STATUS_CHANNEL" default:""`
	LogChan    string `envconfig:"KETI_DISCORD_LOG_CHANNEL" default:""`
}

func (d *discord) BotToken() string {
	if d.Token == "" {
		log.Fatal("Discord Bot token cannot be left blank")
	}

	return fmt.Sprintf("Bot %s", d.Token)
}

var Options EnvConfig
