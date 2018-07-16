package discord

import (
	"encoding/json"

	"github.com/codechimp-io/keti/log"

	"github.com/bwmarrin/discordgo"
)

func (m *Manager) OnDiscordConnected(s *discordgo.Session, evt *discordgo.Connect) {
	m.handleEvent(EventConnected, "")
}

func (m *Manager) OnDiscordDisconnected(s *discordgo.Session, evt *discordgo.Disconnect) {
	m.handleEvent(EventDisconnected, "")
}

func (m *Manager) OnDiscordReady(s *discordgo.Session, evt *discordgo.Ready) {
	m.handleEvent(EventReady, "")
	m.UserID = evt.User.ID
}

func (m *Manager) OnDiscordResumed(s *discordgo.Session, evt *discordgo.Resumed) {
	m.handleEvent(EventResumed, "")
}

func (m *Manager) OnMessageReceive(s *discordgo.Session, e *discordgo.Event) {

	// Ignore events that don't contain data
	if e.Operation != 0 || e.Type == "" {
		return
	}
	// Ignore events in ignoredEventsMap, from ignoreEvents configuration variable
	//		if _, ok := ignoredEventsMap[e.Type]; ok {
	//			return
	//		}

	// Unmarshal event if no data is supplied by DiscordGo

	if e.Struct == nil {
		err := json.Unmarshal(e.RawData, &e.Struct)
		if err != nil {
			log.Warn("Failed to unmarshal event without DiscordGo struct")
		}
	}

	log.Debugf("Type: %s, ShardID: %d, Msg: %v", e.Type, s.ShardID+1, e.Struct)
}

/*
func (m *Manager) OnMessageReceive(s *discordgo.Session, e *discordgo.Event) {

	// Ignore events that don't contain data
	if e.Operation != 0 || e.Type == "" {
		return
	}
	// Ignore events in ignoredEventsMap, from ignoreEvents configuration variable
	//		if _, ok := ignoredEventsMap[e.Type]; ok {
	//			return
	//		}
	// Unmarshal event if no data is supplied by DiscordGo
	if e.Struct == nil {
		err := json.Unmarshal(e.RawData, &e.Struct)
		if err != nil {
			log.Warn("Failed to unmarshal event without DiscordGo struct")
		}
	}

	serializeAndDispatchEvent(discord, stomp, e.Type, e.Struct)
}

*/
