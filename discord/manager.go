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
	Name   string
	UserID string

	// Sessions
	Session       *discordgo.Session
	nsc           *nats.EncodedConn
	eventHandlers []interface{}

	// If set logs connection status events to this channel
	LogChannel string

	// If set keeps an updated satus message in this channel
	StatusMessageChannel string

	// The function that provides the guild counts for this shard, used for the updated status message
	// Should return guilds count
	GuildCountFunc func() int

	// Called on events, by default this is set to a function that logs it to log.Printf
	// You can override this if you want another behaviour, or just set it to nil for nothing.
	OnEvent func(e *Event)

	// SessionFunc creates a new session and returns it, override the default one if you have your own
	// session settings to apply
	SessionFunc SessionFunc

	nextStatusUpdate     time.Time
	statusUpdaterStarted bool

	token string

	//	bareSession *discordgo.Session
	started bool
}

// New creates a new manager with the defaults set, after you have created this you call Manager.Start
// To start connecting
// discord.New("Bot TOKEN", OptLogChannel(someChannel), OptLogEventsToDiscord(true, true))
func New(token string, nsc *nats.EncodedConn) *Manager {
	// Setup defaults
	manager := &Manager{
		token: token,
		nsc:   nsc,
	}

	manager.OnEvent = manager.LogConnectionEventStd
	manager.SessionFunc = manager.StdSessionFunc

	//	manager.bareSession, _ = discordgo.New(token)

	return manager
}

// Adds an event handler
// All event handlers will be added to new sessions automatically.
func (m *Manager) AddHandler(handler interface{}) {
	m.Lock()
	defer m.Unlock()
	m.eventHandlers = append(m.eventHandlers, handler)

	m.Session.AddHandler(handler)
}

// LogConnectionEventStd is the standard connection event logger, it logs it to whatever log.output is set to.
func (m *Manager) LogConnectionEventStd(e *Event) {
	log.Printf(e.String())
}

// StdSessionFunc is the standard session provider, it does nothing to the actual session
func (m *Manager) StdSessionFunc(token string) (*discordgo.Session, error) {
	s, err := discordgo.New(token)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Init will initialize the manager
func (m *Manager) Init() error {
	m.Lock()

	err := m.initSession()
	if err != nil {
		m.Unlock()
		return err
	}

	if !m.statusUpdaterStarted {
		m.statusUpdaterStarted = true
		go m.statusRoutine()
	}

	m.nextStatusUpdate = time.Now()

	m.started = true

	m.Unlock()

	return nil
}

// Start starts the the manager, opening the gateway connection
// this is a blocking call until it exits
func (m *Manager) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	m.Lock()
	if m.Session == nil {
		m.Unlock()
		err := m.Init()
		if err != nil {
			log.Fatalf("Cannot create Discord session: %s", err)
		}
		m.Lock()
	}

	m.Unlock()

	m.Lock()
	err := m.startSession()
	m.Unlock()
	if err != nil {
		log.Fatalf("Cannot start Discord session: %s", err)
	}

	select {
	case <-ctx.Done():
		m.Session.Close()
		log.Info("Discord session closed")
	}
}

// Started determines if the manager have been started
func (m *Manager) Started() bool {
	m.RLock()
	defer m.RUnlock()

	return m.started
}

func (m *Manager) initSession() error {
	session, err := m.SessionFunc(m.token)
	if err != nil {
		return err
	}

	session.AddHandler(m.OnDiscordConnected)
	session.AddHandler(m.OnDiscordDisconnected)
	session.AddHandler(m.OnDiscordReady)
	session.AddHandler(m.OnDiscordResumed)

	// To be removed?
	session.AddHandler(m.OnMessageReceive)

	// Add the user event handlers retroactively
	for _, v := range m.eventHandlers {
		session.AddHandler(v)
	}

	m.Session = session
	return nil
}

func (m *Manager) startSession() error {

	err := m.Session.Open()
	if err != nil {
		return err
	}
	m.handleEvent(EventOpen, "")

	return nil
}

func (m *Manager) statusRoutine() {
	if m.StatusMessageChannel == "" {
		return
	}

	mID := ""

	// Find the initial message id and reuse that message if found
	msgs, err := m.Session.ChannelMessages(m.StatusMessageChannel, 50, "", "", "")
	if err != nil {
		m.handleError(err, "Failed requesting message history in channel")
	} else {
		for _, msg := range msgs {
			// Dunno our own bot id so best we can do is bot
			if msg.Author.ID == m.UserID || len(msg.Embeds) < 1 {
				continue
			}

			nameStr := ""
			if m.Name != "" {
				nameStr = " for " + m.Name
			}

			embed := msg.Embeds[0]
			if embed.Title == "Sharding status"+nameStr {
				// Found it sucessfully
				mID = msg.ID
				break
			}
		}
	}

	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			m.RLock()
			after := time.Now().After(m.nextStatusUpdate)
			m.RUnlock()
			if after {
				m.Lock()
				m.nextStatusUpdate = time.Now().Add(time.Minute)
				m.Unlock()

				nID, err := m.updateStatusMessage(mID)
				if !m.handleError(err, "Failed updating status message") {
					mID = nID
				}
			}
		}
	}
}

func (m *Manager) updateStatusMessage(mID string) (string, error) {
	content := ""

	var numGuilds int
	if m.GuildCountFunc != nil {
		numGuilds = m.GuildCountFunc()
	} else {
		numGuilds = m.StdGuildCountFunc()
	}

	emoji := ""
	if m.Session.DataReady {
		emoji = "👌"
	} else if m.Session != nil {
		emoji = "🕒"
	} else {
		emoji = "🔥"
	}
	content += fmt.Sprintf("[%d/%d]: %s (%d)\n", m.Session.ShardID+1, m.Session.ShardCount, emoji, numGuilds)

	nameStr := ""
	if m.Name != "" {
		nameStr = " for " + m.Name
	}

	embed := &discordgo.MessageEmbed{
		Title:       "Sharding status" + nameStr,
		Description: content,
		Color:       0x4286f4,
		Timestamp:   time.Now().Format(time.RFC3339),
	}

	if mID == "" {
		msg, err := m.Session.ChannelMessageSendEmbed(m.StatusMessageChannel, embed)
		if err != nil {
			return "", err
		}

		return msg.ID, err
	}

	m.Session.UpdateStatus(0, fmt.Sprintf("[%d/%d]: (%d)\n", m.Session.ShardID+1, m.Session.ShardCount, numGuilds))
	_, err := m.Session.ChannelMessageEditEmbed(m.StatusMessageChannel, mID, embed)
	return mID, err
}

func (m *Manager) handleError(err error, msg string) bool {
	if err == nil {
		return false
	}

	m.handleEvent(EventError, msg+": "+err.Error())
	return true
}

func (m *Manager) handleEvent(typ EventType, msg string) {
	if m.OnEvent == nil {
		return
	}

	evt := &Event{
		Type:      typ,
		Shard:     m.Session.ShardID,
		NumShards: m.Session.ShardCount,
		Msg:       msg,
		Time:      time.Now(),
	}

	go m.OnEvent(evt)

	if m.LogChannel != "" {
		go m.logEventToDiscord(evt)
	}

	go func() {
		m.Lock()
		m.nextStatusUpdate = time.Now().Add(time.Second * 2)
		m.Unlock()
	}()
}

// StdGuildCountFunc uses the standard states to return the guilds
func (m *Manager) StdGuildCountFunc() (guilds int) {

	m.RLock()

	m.Session.State.RLock()
	guilds = len(m.Session.State.Guilds)
	m.Session.State.RUnlock()

	m.RUnlock()

	return
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

	_, err := m.Session.ChannelMessageSendEmbed(m.LogChannel, embed)
	m.handleError(err, "Failed sending event to discord")
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

	s := fmt.Sprintf("Discord %s%s", prefix, strings.Title(c.Type.String()))
	if c.Msg != "" {
		s += ": " + c.Msg
	}

	return s
}

// Event holds data for an event
type NatsEvent struct {
	Type      string
	Shard     int
	NumShards int
	Data      interface{}
	// When this event occured
	Time time.Time
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
