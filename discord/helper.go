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
	mgr.ShardsCount = config.Options.Discord.ShardCount
	mgr.ShardsOffset = config.Options.Discord.ShardOffset
	mgr.ShardsTotal = config.Options.Discord.ShardTotal

	wg.Add(1)
	go mgr.Start(ctx, wg)

	if len(mgr.Sessions) > 0 {
		for {
			if mgr.Started() {
				break
			}

			time.Sleep(100 * time.Millisecond)
		}
	}

	log.Info("Connected to Discord")
}
