package server

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/baconstrip/kiken/message"
	"golang.org/x/net/websocket"
)

type SessionID uint64

type SessionManager struct {
	mu sync.RWMutex

	sessions map[SessionID]SessionVar
	names    map[string]SessionID

	connections     map[SessionID]*Connection
	recentlyDropped map[SessionID]time.Time
}

type Connection struct {
	in  chan message.ClientMessage
	out chan message.ServerMessage
	soc *websocket.Conn
}

type SessionVar struct {
	name     string
	passcode string
	host     bool
}

// createSession generates a random sessionID for a user and stores the vars
// in an association to that ID. It writes the cookie the client needs to the
// ResponseWriter passed as w.
func (s *SessionManager) createSession(vars SessionVar, w http.ResponseWriter) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var max big.Int
	max.SetUint64(maxUint64)
	keyBig, err := rand.Int(rand.Reader, &max)
	if err != nil {
		return fmt.Errorf("failed to generate random number: %v", err)
	}

	key := SessionID(keyBig.Uint64())

	s.sessions[key] = vars
	s.names[vars.name] = key
	cookie := http.Cookie{
		Name:    sessionName,
		Value:   strconv.FormatUint(uint64(key), 10),
		Expires: time.Now().Add(time.Hour * 24),
	}
	http.SetCookie(w, &cookie)
	return nil
}

// Returns the SessionID that corresponds to the given name, or the boolean set
// to false.
func (s *SessionManager) IDFromName(name string) (SessionID, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.names[name]
	return session, ok
}

// destroySession removes all the information associated with a session,
// including the variables and connections. Should only be called after
// dropConnection.
func (s *SessionManager) destroySession(key SessionID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.names, s.sessions[key].name)
	delete(s.sessions, key)
}

func (s *SessionManager) addConnection(id SessionID, ws *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.connections[id] = &Connection{
		soc: ws,
		in:  make(chan message.ClientMessage, 1000),
		out: make(chan message.ServerMessage, 1000),
	}
	delete(s.recentlyDropped, id)
}

// dropConnection removes the copy of the message queues and socket associated
// with a user from the server. It returns three values, the name of the player
// leaving, whether or not this is handler has been called for this player
// recently, and finally whether or not this player was a host.
func (s *SessionManager) dropConnection(id SessionID) (string, bool, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := s.sessions[id].name
	host := s.sessions[id].host

	if c, ok := s.connections[id]; ok {
		close(c.in)
		close(c.out)
	}

	delete(s.connections, id)

	retVal := true

	// If they've been dropped already in the past ten seconds without
	// rejoining, this is probably a duplicate.
	if t, ok := s.recentlyDropped[id]; ok && time.Since(t) < 10*time.Second {
		retVal = false
	}
	s.recentlyDropped[id] = time.Now()

	return name, retVal, host
}

// withConnection locks the mutex and runs an operation. Returns an error if
// the connection isn't found. Will also return an error if f results in an
// error.
func (s *SessionManager) withConnection(id SessionID, f func(*Connection) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.connections[id]
	if !ok {
		return fmt.Errorf("failed to run withConnection(): connection not found for id %v", id)
	}
	return f(c)
}

// writeMessage schedules a message to be sent to a client asynchronously.
// Do not modify message after scheduling. Returns an error if the client is
// already disconnected.
func (s *SessionManager) writeMessage(id SessionID, msg message.ServerMessage) error {
	return s.withConnection(id, func(c *Connection) error {
		c.out <- msg
		return nil
	})
}

// messageAll schedules a message to be sent asynchronously to all clients.
func (s *SessionManager) messageAll(msg message.ServerMessage) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for id := range s.connections {
		s.writeMessage(id, msg)
	}
}

// messageHost schedules a message to be sent asynchronously to the host.
func (s *SessionManager) messageHost(msg message.ServerMessage) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for id := range s.connections {
		if s.sessions[id].host {
			s.writeMessage(id, msg)
		}
	}
}

// messagePlayers schedules a message to be sent asynchronously to all players
// except the host.
func (s *SessionManager) messagePlayers(msg message.ServerMessage) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for id := range s.connections {
		if !s.sessions[id].host {
			s.writeMessage(id, msg)
		}
	}
}

// messagePlayer schedules a message to be sent asynchronously to a single
// player, whose name is given.
func (s *SessionManager) messagePlayer(msg message.ServerMessage, name string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for id := range s.connections {
		if s.sessions[id].name == name {
			s.writeMessage(id, msg)
			return
		}
	}
}

func (s *SessionManager) userExists(name string, caseInsensitive bool) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, vars := range s.sessions {
		if vars.name == name || (strings.EqualFold(name, vars.name) && caseInsensitive) {
			return true
		}
	}
	return false
}

func (s *SessionManager) correctPasscode(name string, caseInsensitive bool, passcode string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, vars := range s.sessions {
		if vars.name == name || (strings.EqualFold(name, vars.name) && caseInsensitive) {
			return vars.passcode == passcode
		}
	}
	return false
}
