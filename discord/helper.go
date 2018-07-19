package discord

import (
	"context"
	"sync"
	"time"

	"github.com/codechimp-io/keti/config"
	"github.com/codechimp-io/keti/log"
	"github.com/codechimp-io/keti/version"

	"github.com/nats-io/go-nats"
)

// Run starts new Discord manager.
func Run(ctx context.Context, wg *sync.WaitGroup, nsc *nats.EncodedConn) {

	// Configure new manager
	mgr := New(config.Options.Discord.BotToken(), nsc)
	mgr.Name = version.Name
	mgr.LogChannel = "466629625167085571"
	mgr.ShardID = config.Options.Discord.ShardID
	mgr.ShardCount = config.Options.Discord.ShardCount

	wg.Add(1)
	go mgr.Start(ctx, wg)

	if mgr.Session != nil {
		for {
			if mgr.Started() {
				break
			}

			time.Sleep(100 * time.Millisecond)
		}
	}

	log.Info("Connected to Discord")
}
