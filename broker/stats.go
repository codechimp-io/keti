// Export NATS stats as Prometheus metrics
package broker

import (
	"context"
	"time"

	"github.com/codechimp-io/keti/log"
	"github.com/codechimp-io/keti/version"

	gnatsd "github.com/nats-io/gnatsd/server"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	connectionsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gnatsd_network_connections",
		Help: "Current connections on the broker",
	}, []string{"identity"})

	totalConnectionsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gnatsd_network_total_connections",
		Help: "Total connections received since start",
	}, []string{"identity"})

	routesGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gnatsd_network_routes",
		Help: "Current active routes to other brokers",
	}, []string{"identity"})

	remotesGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gnatsd_network_remotes",
		Help: "Current active connections to other brokers",
	}, []string{"identity"})

	inMsgsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gnatsd_network_in_msgs",
		Help: "Messages received by the broker",
	}, []string{"identity"})

	outMsgsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gnatsd_network_out_msgs",
		Help: "Messages sent by the broker",
	}, []string{"identity"})

	inBytesGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gnatsd_network_in_bytes",
		Help: "Total size of messages received by the broker",
	}, []string{"identity"})

	outBytesGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gnatsd_network_out_bytes",
		Help: "Total size of messages sent by the broker",
	}, []string{"identity"})

	slowConsumerGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gnatsd_network_slow_consumers",
		Help: "Total number of clients who were considered slow consumers",
	}, []string{"identity"})

	subscriptionsGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gnatsd_network_subscriptions",
		Help: "Number of active subscriptions to subjects on this broker",
	}, []string{"identity"})
)

func init() {
	prometheus.MustRegister(connectionsGauge)
	prometheus.MustRegister(totalConnectionsGauge)
	prometheus.MustRegister(routesGauge)
	prometheus.MustRegister(remotesGauge)
	prometheus.MustRegister(inMsgsGauge)
	prometheus.MustRegister(outMsgsGauge)
	prometheus.MustRegister(inBytesGauge)
	prometheus.MustRegister(outBytesGauge)
	prometheus.MustRegister(slowConsumerGauge)
	prometheus.MustRegister(subscriptionsGauge)
}

func (s *Server) scrapeVarz() (*gnatsd.Varz, error) {
	return s.Server.Varz(&gnatsd.VarzOptions{})
}

func (s *Server) publishStats(ctx context.Context, interval time.Duration) {
	if s.Opts.HTTPPort == 0 {
		return
	}

	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			log.Debug("Updating Prometheus stats from NATS /varz")

			s.updatePrometheus()
		case <-ctx.Done():
			return
		}
	}
}

func (s *Server) updatePrometheus() {
	varz, err := s.scrapeVarz()
	if err != nil {
		log.Errorf("Could not update stats from NATS /varz: %s", err)
		return
	}

	i := version.Name

	connectionsGauge.WithLabelValues(i).Set(float64(varz.Connections))
	totalConnectionsGauge.WithLabelValues(i).Set(float64(varz.TotalConnections))
	routesGauge.WithLabelValues(i).Set(float64(varz.Routes))
	remotesGauge.WithLabelValues(i).Set(float64(varz.Remotes))
	inMsgsGauge.WithLabelValues(i).Set(float64(varz.InMsgs))
	outMsgsGauge.WithLabelValues(i).Set(float64(varz.OutMsgs))
	inBytesGauge.WithLabelValues(i).Set(float64(varz.InBytes))
	outBytesGauge.WithLabelValues(i).Set(float64(varz.OutBytes))
	slowConsumerGauge.WithLabelValues(i).Set(float64(varz.SlowConsumers))
	subscriptionsGauge.WithLabelValues(i).Set(float64(varz.Subscriptions))
}
