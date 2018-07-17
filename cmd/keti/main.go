package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/codechimp-io/keti/broker"
	"github.com/codechimp-io/keti/discord"
	"github.com/codechimp-io/keti/log"
	"github.com/codechimp-io/keti/version"
)

var (
	ctx    context.Context
	cancel func()
	wg     *sync.WaitGroup
	err    error
)

func main() {

	// Set producer name in logs
	log.WithCaller(version.Name)

	wg = &sync.WaitGroup{}

	log.Infof("Starting %s", version.Info())

	// Init context
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	// Run the embeded broker and obtain connection
	nc := broker.RunAndConnect(ctx, wg, false)

	// Run discord manager
	discord.Run(ctx, wg, nc, "Bot NDI5MzAwNDU5MTE1OTA1MDI0.Die2kg.YRx6kqnONTV_EULIBG3E6iOJ0V0")

	// Spawn OS Signal watcher
	signalWatcher()

	// Wait for all processes to finish
	wg.Wait()
}

func signalWatcher() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	for {
		select {
		case sig := <-sigs:
			switch sig {
			case syscall.SIGINT:
				log.Info("Shutdown requested with CTRL+C!")
				cancel()
			case syscall.SIGTERM:
				log.Info("Shutdown requested with SIGTERM!")
				cancel()
			case syscall.SIGQUIT:
				log.Info("Shutdown requested with SIGQUIT!")
				cancel()
			}
		case <-ctx.Done():
			//			log.Warn("Server Exiting...")
			return
		}
	}
}
