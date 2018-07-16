package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/codechimp-io/keti/broker"
	//	"github.com/codechimp-io/keti/version"
	"github.com/codechimp-io/keti/log"
)

const (
	// NAME is the name of the daemon
	NAME = "keti"
)

var (
	ctx    context.Context
	cancel func()
	wg     *sync.WaitGroup
	err    error
)

func main() {

	// Set producer name in logs
	log.WithCaller(NAME)

	wg = &sync.WaitGroup{}

	log.Info("Starting...")

	// Init context
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	// Run the embeded broker and obtain connection
	nc := broker.RunAndConnect(ctx, wg, false)

	// Simple Async Subscriber
	nc.Subscribe("foo", func(s string) {
    	log.Printf("Received a message: %s\n", s)
	})

	// Simple Publisher
	nc.Publish("foo", "Hello World")

	// Spawn OS Signal watcher
	signalWatcher()

	// Wait for all processes to finish
	wg.Wait()
}

func signalWatcher() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	for {
		select {
		case <-sig:
			log.Warnf("Shutting down on %s", sig)
			cancel()
		case <-ctx.Done():
			log.Warn("Exiting...")
			return
		}
	}
}
