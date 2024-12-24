package editor

import (
	"sync"

	"github.com/baconstrip/kiken/server"
)

var DataDir string

type EditorDriver struct {
	mu *sync.RWMutex

	server *server.Server

	// sessions contains a mapping between names and the editor sessions
	sessions map[string]EditorSession
}

type EditorSession struct {
	currentShow *Show
}

func NewEditorDriver(s *server.Server) *EditorDriver {
	return &EditorDriver{
		mu:       &sync.RWMutex{},
		server:   s,
		sessions: make(map[string]EditorSession),
	}
}
