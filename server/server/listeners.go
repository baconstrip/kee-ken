package server

import (
	"log"
	"sync"

	"github.com/baconstrip/kiken/message"
)

type JoinListener func(name string, host bool) error
type LeaveListener func(name string, host bool) error
type ClientMessageListener func(name string, host bool, msg message.ClientMessage) error

type ListenerManager struct {
	mu sync.RWMutex

	joinListeners  []JoinListener
	leaveListeners []LeaveListener
	msgListeners   map[string][]ClientMessageListener
}

func NewListenerManager() *ListenerManager {
	return &ListenerManager{
		msgListeners: make(map[string][]ClientMessageListener),
	}
}

// RegisterJoin adds a JoinListener that is called when a client connects.
func (l *ListenerManager) RegisterJoin(j JoinListener) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.joinListeners = append(l.joinListeners, j)
}

// RegisterLeave adds a LeaveListener that is called when a client discconects.
func (l *ListenerManager) RegisterLeave(ls LeaveListener) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.leaveListeners = append(l.leaveListeners, ls)
}

func (l *ListenerManager) ClearListeners() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.joinListeners = nil
	l.leaveListeners = nil
	l.msgListeners = nil
}

// RegisterMessage adds a ClientMessageListener for the given message type,
// must be one of the names of a client message in the messages package.
func (l *ListenerManager) RegisterMessage(messageType string, c ClientMessageListener) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.msgListeners[messageType] = append(l.msgListeners[messageType], c)
}

func (l *ListenerManager) dispatchJoin(name string, host bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, listener := range l.joinListeners {
		listenerCpy := listener
		go func() {
			if err := listenerCpy(name, host); err != nil {
				log.Printf("Error processing listener for join: %v", err)
			}
		}()
	}
}

func (l *ListenerManager) dispatchLeave(name string, host bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, listener := range l.leaveListeners {
		listenerCpy := listener
		go func() {
			if err := listenerCpy(name, host); err != nil {
				log.Printf("Error processing listener for leave: %v", err)
			}
		}()
	}
}

func (l *ListenerManager) dispatchMessage(name string, host bool, msg message.ClientMessage) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for _, listener := range l.msgListeners[msg.Type] {
		listenerCpy := listener
		go func() {
			if err := listenerCpy(name, host, msg); err != nil {
				log.Printf("Error processing listener for message %v: %v", msg.Type, err)
			}
		}()
	}
}
