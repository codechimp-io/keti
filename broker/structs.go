package broker

import "time"

// GatewayEvent holds data for an event sent from the gateway
type GatewayEvent struct {
	Shard     int
	NumShards int
	UserID    interface{}
	Data      interface{}
	// When this event occured
	Time time.Time
}
