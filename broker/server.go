package broker

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/codechimp-io/keti/log"

	gnatsd "github.com/nats-io/gnatsd/server"
)

// Server wrapper around NATS Server
type Server struct {
	Server *gnatsd.Server
	Opts   *gnatsd.Options

	started bool

	mu *sync.Mutex
}

// New creates a new instance of the Server struct with a fully configured NATS embedded server
func NewServer(debug bool) (s *Server, err error) {
	s = &Server{
		Opts:    &gnatsd.Options{},
		started: false,
		mu:      &sync.Mutex{},
	}

	s.Opts.Host = gnatsd.DEFAULT_HOST
	s.Opts.Port = gnatsd.DEFAULT_PORT
	s.Opts.Logtime = false
	s.Opts.MaxConn = gnatsd.DEFAULT_MAX_CONNECTIONS
	s.Opts.WriteDeadline = gnatsd.DEFAULT_FLUSH_DEADLINE
	s.Opts.NoSigs = true

	if debug {
		s.Opts.Debug = true
	}

	s.Opts.HTTPHost = gnatsd.DEFAULT_HOST
	s.Opts.HTTPPort = gnatsd.DEFAULT_HTTP_PORT

	// Configure cluster options
	err = s.configureCluster()
	if err != nil {
		return s, fmt.Errorf("Could not configure NATS Cluster: %s", err)
	}

	s.Server = gnatsd.New(s.Opts)

	// Setup custom logger
	s.Server.SetLogger(log.NewNATSLogger(), debug, false)

	return
}

func (s *Server) configureCluster() (err error) {
	s.Opts.Cluster.Host = gnatsd.DEFAULT_HOST
	s.Opts.Cluster.NoAdvertise = true
	s.Opts.Cluster.Port = gnatsd.DEFAULT_PORT + 1000
	s.Opts.Cluster.Username = "clusterino"
	s.Opts.Cluster.Password = "s3cret"
	/*
		peers, err := config.GetPeers()
		if err != nil {
			return fmt.Errorf("Could not determine network broker peers: %s", err)
		}

		for _, p := range peers {
			u, err := p.URL()
			if err != nil {
				return fmt.Errorf("Could not parse Peer configuration: %s", err)
			}

			log.Infof("Adding %s as network peer", u.String())
			s.opts.Routes = append(s.opts.Routes, u)
		}
	*/
	// Remove any host/ip that points to itself in Route
	newroutes, err := gnatsd.RemoveSelfReference(s.Opts.Cluster.Port, s.Opts.Routes)
	if err != nil {
		return fmt.Errorf("Could not remove own Self from cluster configuration: %s", err)
	}

	s.Opts.Routes = newroutes

	return
}

// HTTPHandler Exposes the gnatsd HTTP Handler
func (s *Server) HTTPHandler() http.Handler {
	return s.Server.HTTPHandler()
}

// Start the embedded NATS instance, this is a blocking call until it exits
func (s *Server) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	go s.Server.Start()

	s.mu.Lock()
	s.started = true
	s.mu.Unlock()

	s.publishStats(ctx, 10*time.Second)

	select {
	case <-ctx.Done():
		s.Server.Shutdown()
		log.Info("NATS Broker stopped")
	}
}

// Started determines if the server have been started
func (s *Server) Started() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.started
}
