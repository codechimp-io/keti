package broker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/codechimp-io/keti/log"

	"github.com/nats-io/go-nats"
)

// RunAndConnect starts new embedded NATS instance and returns JSON encoded connection to it.
func RunAndConnect(ctx context.Context, wg *sync.WaitGroup, debug bool) *nats.EncodedConn {

	// Configure new embed broker
	nsq, err := NewServer(debug)
	if err != nil {
		log.Fatalf("Cannot configure new NATS Broker %s", err)
	}

	wg.Add(1)
	go nsq.Start(ctx, wg)

	if nsq.Server != nil {
		for {
			if nsq.Started() {
				break
			}

			time.Sleep(100 * time.Millisecond)
		}
	}

	opts := nats.Options{}
    opts.Name = "keti"
	opts.Url = fmt.Sprintf("nats://%s:%v", nsq.Opts.Host, nsq.Opts.Port)

	nc, err := NewEncodedClient(&opts)
	if err != nil {
	        log.Fatalf("Error connecting to local NATS server: %s", err)
	}

	log.Infof("Connected to NATS: %s", opts.Url)

	return nc
}
