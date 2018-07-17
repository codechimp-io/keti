package discord

import (
	"context"
	"sync"
	"time"

	"github.com/codechimp-io/keti/log"
	"github.com/codechimp-io/keti/version"

	"github.com/nats-io/go-nats"
)

// Run starts new Discord manager.
func Run(ctx context.Context, wg *sync.WaitGroup, nsc *nats.EncodedConn, token string) {

	// Configure new manager
	mgr := New(token, nsc)
	mgr.Name = version.Name
	mgr.LogChannel = "466629625167085571"
	mgr.StatusMessageChannel = "468467517061464076"

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