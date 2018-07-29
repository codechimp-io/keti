package discord

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/codechimp-io/keti/log"

	"github.com/bwmarrin/discordgo"
	"github.com/nats-io/go-nats"
)

type SessionFunc func(token string) (*discordgo.Session, error)

// Manager implements
type Manager struct {
	sync.RWMutex

	// Name of the bot, to appear in the title of the updated status message
	Name string

	// Sessions
	Sessions map[int]*discordgo.Session
	nsc      *nats.EncodedConn

	// handlers
	eventHandlers []interface{}

	// If set logs connection status events to this channel
	LogChannel string

	// The function that provides the guild counts for this shard, used for the updated status message
	// Should return guilds count
	GuildCountFunc func() int

	// Called on events, by default this is set to a function that logs it to log.Printf
	// You can override this if you want another behaviour, or just set it to nil for nothing.
	OnEvent func(e *Event)

	// SessionFunc creates a new session and returns it, override the default one if you have your own
	// session settings to apply
	SessionFunc SessionFunc

	// Total Shards and current number of shards for this instance
	ShardsTotal  int
	ShardsCount  int
	ShardsOffset int

	token string

	bareSession *discordgo.Session
	started     bool
}

// New creates a new shard manager with the defaults set, after you have created this you call Manager.Start
// To start connecting
// dshardmanager.New("Bot asd", OptLogChannel(someChannel), OptLogEventsToDiscord(true, true))
func New(token string, nsc *nats.EncodedConn) *Manager {
	// Setup defaults
	manager := &Manager{
		token:       token,
		ShardsCount: -1,
		nsc:         nsc,
	}

	manager.OnEvent = manager.LogConnectionEventStd
	manager.SessionFunc = manager.DefaultSessionFunc

	manager.bareSession, _ = discordgo.New(token)

	return manager
}

// GetNumShards returns the current set number of shards for this instance
func (m *Manager) GetNumShards() int {
	return m.ShardsCount
}

// Adds an event handler to all shards
// All event handlers will be added to new sessions automatically.
func (m *Manager) AddHandler(handler interface{}) {
	m.Lock()
	defer m.Unlock()
	m.eventHandlers = append(m.eventHandlers, handler)

	if len(m.Sessions) > 0 {
		for _, v := range m.Sessions {
			v.AddHandler(handler)
		}
	}
}

// Init initializesthe manager, retreiving the recommended shard count if needed
// and initalizes all the shards
func (m *Manager) Init() error {
	m.Lock()

	// If no sharding is set default to one shard
	if m.ShardsCount < 1 {
		m.ShardsCount = 1
		m.ShardsTotal = 1
		m.ShardsOffset = 0
	}

	m.Sessions = make(map[int]*discordgo.Session, m.ShardsCount)
	for i := m.ShardsOffset; i < m.ShardsCount; i++ {
		err := m.initSession(i)
		if err != nil {
			m.Unlock()
			return err
		}
	}

	m.Unlock()

	return nil
}

// Start starts the the manager, opening the gateway connection
// this is a blocking call until it exits
func (m *Manager) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	m.Lock()
	if m.Sessions == nil {
		m.Unlock()
		err := m.Init()
		if err != nil {
			log.Fatalf("Cannot create Discord session: %s", err)
		}
		m.Lock()
	}

	m.Unlock()

	for i := m.ShardsOffset; i < m.ShardsCount; i++ {
		if i != 0 {
			// One indentify every 5 seconds
			time.Sleep(time.Second * 5)
		}

		m.Lock()
		err := m.startSession(i)
		m.Unlock()
		if err != nil {
			log.Fatalf("Cannot start Discord ShardID: %d, session: %s", err, i)
		}
	}

	select {
	case <-ctx.Done():
		m.StopAll()
		log.Info("Discord sessions closed")
	}

}

// StopAll stops all the shard sessions and returns the last error that occured
func (m *Manager) StopAll() (err error) {
	m.Lock()
	for _, v := range m.Sessions {
		if e := v.Close(); e != nil {
			err = e
		}
	}
	m.Unlock()

	return
}

// Started determines if the manager have been started
func (m *Manager) Started() bool {
	m.RLock()
	defer m.RUnlock()

	return m.started
}

func (m *Manager) initSession(shard int) error {
	session, err := m.SessionFunc(m.token)
	if err != nil {
		return err
	}

	session.ShardCount = m.ShardsTotal
	session.ShardID = shard

	session.AddHandler(m.OnDiscordConnected)
	session.AddHandler(m.OnDiscordDisconnected)
	session.AddHandler(m.OnDiscordReady)
	session.AddHandler(m.OnDiscordResumed)
	session.AddHandler(m.OnDiscordEvent)

	// Add the user event handlers retroactively
	for _, v := range m.eventHandlers {
		session.AddHandler(v)
	}

	m.Sessions[shard] = session
	return nil
}

func (m *Manager) startSession(shard int) error {

	err := m.Sessions[shard].Open()
	if err != nil {
		return err
	}
	m.handleEvent(EventOpen, shard, "")

	return nil
}

// Session retrieves a session from the sessions map, rlocking it in the process
func (m *Manager) Session(shardID int) *discordgo.Session {
	m.RLock()
	defer m.RUnlock()
	return m.Sessions[m.ShardsOffset+shardID]
}

// LogConnectionEventStd is the standard connection event logger, it logs it to whatever log.output is set to.
func (m *Manager) LogConnectionEventStd(e *Event) {
	log.Printf("[Shard Manager] %s", e.String())
}

func (m *Manager) handleError(err error, shard int, msg string) bool {
	if err == nil {
		return false
	}

	m.handleEvent(EventError, shard, msg+": "+err.Error())
	return true
}

func (m *Manager) handleEvent(typ EventType, shard int, msg string) {
	if m.OnEvent == nil {
		return
	}

	evt := &Event{
		Type:      typ,
		Shard:     shard,
		NumShards: m.ShardsTotal,
		Msg:       msg,
		Time:      time.Now(),
	}

	m.OnEvent(evt)

	if m.LogChannel != "" {
		go m.logEventToDiscord(evt)
	}
}

// DefaultSessionFunc is the default session provider, it does nothing to the actual session
func (m *Manager) DefaultSessionFunc(token string) (*discordgo.Session, error) {
	s, err := discordgo.New(token)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (m *Manager) logEventToDiscord(evt *Event) {
	if evt.Type == EventError {
		return
	}

	prefix := ""
	if m.Name != "" {
		prefix = m.Name + ": "
	}

	str := evt.String()
	embed := &discordgo.MessageEmbed{
		Description: prefix + str,
		Timestamp:   evt.Time.Format(time.RFC3339),
		Color:       eventColors[evt.Type],
	}

	_, err := m.bareSession.ChannelMessageSendEmbed(m.LogChannel, embed)
	m.handleError(err, evt.Shard, "Failed sending event to discord")
}

// Event holds data for an event
type Event struct {
	Type EventType

	Shard     int
	NumShards int

	Msg string

	// When this event occured
	Time time.Time
}

func (c *Event) String() string {
	prefix := ""
	if c.Shard > -1 {
		prefix = fmt.Sprintf("[%d/%d] ", c.Shard+1, c.NumShards)
	}

	s := fmt.Sprintf("%s%s", prefix, strings.Title(c.Type.String()))
	if c.Msg != "" {
		s += ": " + c.Msg
	}

	return s
}

type EventType int

const (
	// Sent when the connection to the gateway was established
	EventConnected EventType = iota

	// Sent when the connection is lose
	EventDisconnected

	// Sent when the connection was sucessfully resumed
	EventResumed

	// Sent on ready
	EventReady

	// Sent when Open() is called
	EventOpen

	// Sent when Close() is called
	EventClose

	// Sent when an error occurs
	EventError
)

var (
	eventStrings = map[EventType]string{
		EventOpen:         "opened",
		EventClose:        "closed",
		EventConnected:    "connected",
		EventDisconnected: "disconnected",
		EventResumed:      "resumed",
		EventReady:        "ready",
		EventError:        "error",
	}

	eventColors = map[EventType]int{
		EventOpen:         0xec58fc,
		EventClose:        0xff7621,
		EventConnected:    0x54d646,
		EventDisconnected: 0xcc2424,
		EventResumed:      0x5985ff,
		EventReady:        0x00ffbf,
		EventError:        0x7a1bad,
	}
)

func (c EventType) String() string {
	return eventStrings[c]
}
