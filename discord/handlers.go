package discord

import (
	"encoding/json"
	"time"

	"github.com/codechimp-io/keti/log"

	"github.com/bwmarrin/discordgo"
)

var allowedEventsMap = map[string]struct{}{
	"CHANNEL_CREATE": struct{}{},
	"CHANNEL_DELETE": struct{}{},
	//	"CHANNEL_PINS_UPDATE": struct{}{},
	"CHANNEL_UPDATE":      struct{}{},
	"GUILD_BAN_ADD":       struct{}{},
	"GUILD_BAN_REMOVE":    struct{}{},
	"GUILD_CREATE":        struct{}{},
	"GUILD_DELETE":        struct{}{},
	"GUILD_UPDATE":        struct{}{},
	"GUILD_MEMBER_ADD":    struct{}{},
	"GUILD_MEMBER_REMOVE": struct{}{},
	"GUILD_MEMBER_UPDATE": struct{}{},
	"MESSAGE_CREATE":      struct{}{},
	"PRESENCE_UPDATE":     struct{}{},
}

func (m *Manager) OnDiscordConnected(s *discordgo.Session, e *discordgo.Connect) {
	m.handleEvent(EventConnected, "")
}

func (m *Manager) OnDiscordDisconnected(s *discordgo.Session, e *discordgo.Disconnect) {
	m.handleEvent(EventDisconnected, "")
}

func (m *Manager) OnDiscordReady(s *discordgo.Session, e *discordgo.Ready) {
	// Set self ID
	m.UserID = e.User.ID
	m.Session.StateEnabled = false

	m.handleEvent(EventReady, "")
}

func (m *Manager) OnDiscordResumed(s *discordgo.Session, evt *discordgo.Resumed) {
	m.handleEvent(EventResumed, "")
}

func (m *Manager) OnEventReceive(s *discordgo.Session, e *discordgo.Event) {

	// Ignore events that don't contain data
	if e.Operation != 0 || e.Type == "" {
		return
	}

	// Ignore events in ignoredEventsMap
	if _, ok := allowedEventsMap[e.Type]; !ok {
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
	evt := &NatsEvent{
		UserID:    s.State.User.ID,
		Shard:     m.Session.ShardID + 1,
		NumShards: m.Session.ShardCount,
		Data:      e,
		Time:      time.Now(),
	}

	// Publish message
	m.nsc.Publish("gateway:incomming", evt)

	//	log.Debugf("Type: %s, ShardID: %d, Msg: %s", e.Type, s.ShardID+1, e.RawData)
}
