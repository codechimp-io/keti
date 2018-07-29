package discord

import (
	"encoding/json"
	"time"

	"github.com/codechimp-io/keti/broker"
	"github.com/codechimp-io/keti/log"

	"github.com/bwmarrin/discordgo"
)

var ignoredEventsMap = map[string]struct{}{
	"CHANNEL_PINS_UPDATE": struct{}{},
	"MESSAGE_UPDATE":      struct{}{},
	"TYPING_START":        struct{}{},
}

func (m *Manager) OnDiscordConnected(s *discordgo.Session, e *discordgo.Connect) {
	m.handleEvent(EventConnected, s.ShardID, "")
}

func (m *Manager) OnDiscordDisconnected(s *discordgo.Session, e *discordgo.Disconnect) {
	m.handleEvent(EventDisconnected, s.ShardID, "")
}

func (m *Manager) OnDiscordReady(s *discordgo.Session, e *discordgo.Ready) {
	// Disable State cache
	m.Sessions[s.ShardID].StateEnabled = false

	m.handleEvent(EventReady, s.ShardID, "")
}

func (m *Manager) OnDiscordResumed(s *discordgo.Session, evt *discordgo.Resumed) {
	m.handleEvent(EventResumed, s.ShardID, "")
}

func (m *Manager) OnDiscordEvent(s *discordgo.Session, e *discordgo.Event) {

	// Ignore events that don't contain data
	if e.Operation != 0 || e.Type == "" {
		return
	}

	// Ignore events in ignoredEventsMap
	if _, ok := ignoredEventsMap[e.Type]; ok {
		return
	}

	// Unmarshal event if no data is supplied by DiscordGo
	if e.Struct == nil {
		err := json.Unmarshal(e.RawData, &e.Struct)
		if err != nil {
			log.Warn("Failed to unmarshal event without DiscordGo struct")
		}
	}

	// Create NATS messaage and send
	evt := &broker.GatewayEvent{
		Shard:  s.ShardID,
		UserID: s.State.User.ID,
		Data:   e,
		Time:   time.Now(),
	}

	// Publish message
	m.nsc.Publish("gateway:exchange", evt)

	//	log.Debugf("Type: %s, ShardID: %d, Msg: %s", e.Type, s.ShardID+1, e.RawData)
}
