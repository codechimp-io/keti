// Invite link: https://discordapp.com/api/oauth2/authorize?client_id=468814860596019201&permissions=271641727&scope=bot
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
	// Init sync.WaitGroup and Context
	wg = &sync.WaitGroup{}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	log.Infof("Starting %s", version.Info())

	// Run the embeded broker and obtain connection
	nc := broker.RunAndConnect(ctx, wg)

	// Run discord manager
	discord.Run(ctx, wg, nc)

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
