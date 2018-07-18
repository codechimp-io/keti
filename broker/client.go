package broker

import (
	"errors"
	"time"

	"github.com/codechimp-io/keti/log"

	"github.com/nats-io/go-nats"
)

// DefaultNatsEstablishTimeout is the default timeout for
// calls to EstablishNatsConnection.
var DefaultEstablishTimeout = 60 * time.Second

// NewConnection creates named connection to NATS server,
// it returns connection and any errors encountered.
func NewClient(o *nats.Options) (*nats.Conn, error) {
	opts := nats.Options{
		Url:            o.Url,
		Name:           o.Name,
		AllowReconnect: true,
		MaxReconnect:   10,
		ReconnectWait:  5 * time.Second,
		Timeout:        1 * time.Second,
		ClosedCB: func(conn *nats.Conn) {
			log.Info("NATS Client Connection Closed")
		},
		DisconnectedCB: func(conn *nats.Conn) {
			log.Info("NATS Client Disconnected")
		},
		ReconnectedCB: func(conn *nats.Conn) {
			log.Info("NATS Client Reconnected")
		},
		AsyncErrorCB: func(conn *nats.Conn, sub *nats.Subscription, err error) {
			log.Errorf("NATS async error for %s: %s", sub, err)
		},
	}

	nc, err := opts.Connect()
	if err != nil {
		return nil, err
	}

	return nc, nil
}

// NewEncodedClient is a blocking way to create and establish connection
// to the NATS server. The function will only return after a timeout
// has reached or a connection has been established. It returns
// the connection and any timeout error encountered.
func NewEncodedClient(o *nats.Options) (*nats.EncodedConn, error) {
	if o.Timeout == 0 {
		o.Timeout = DefaultEstablishTimeout
	}
	connch := make(chan *nats.Conn, 1)
	errch := make(chan error, 1)
	go func() {
		notify := true
		for {
			nc, err := NewClient(o)
			if err == nil {
				connch <- nc
				break
			}
			switch err {
			case nats.ErrTimeout:
				fallthrough
			case nats.ErrNoServers:
				if notify {
					notify = false
					log.Error("Waiting for NATS server to become available")
				}
				time.Sleep(1 * time.Second)
				continue
			default:
				errch <- err
				break
			}
		}
	}()

	select {
	case conn := <-connch:
		ec, err := nats.NewEncodedConn(conn, nats.JSON_ENCODER)
		if err != nil {
			return nil, err
		}
		return ec, nil
	case err := <-errch:
		return nil, err
	case <-time.After(o.Timeout):
		return nil, errors.New("NATS connection: timeout")
	}
}
