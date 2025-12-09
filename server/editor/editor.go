package editor

import (
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"path"
	"sync"

	"github.com/baconstrip/kiken/message"
	"github.com/baconstrip/kiken/server"
	"github.com/baconstrip/kiken/util"
)

var DataDir string

type EditorDriver struct {
	mu *sync.RWMutex

	server         *server.Server
	editorListener *server.ListenerManager

	// sessions contains a mapping between names and the editor sessions
	sessions map[string]*EditorSession

	// Contains a mapping between show IDs and the full filename for a show
	knownShows map[string]string
}

type EditorSession struct {
	currentShow *Show
}

func NewEditorDriver(s *server.Server, editorListener *server.ListenerManager) *EditorDriver {
	return &EditorDriver{
		mu:             &sync.RWMutex{},
		server:         s,
		sessions:       make(map[string]*EditorSession),
		editorListener: editorListener,
	}
}

func (e *EditorDriver) Start() {
	e.editorListener.RegisterJoin(e.onJoinManageEditor)
	e.editorListener.RegisterMessage("RequestShow", e.onRequestShowPresentShows)
	e.editorListener.RegisterMessage("SelectShow", e.onSelectShowActivateShow)

	e.refreshGamesFromDisk()
}

func (e *EditorDriver) onRequestShowPresentShows(name string, _ bool, _ message.ClientMessage) error {
	shows := make(map[string]string)

	for id, f := range e.knownShows {
		ext := path.Ext(f)
		cleaned := f[:len(f)-len(ext)]

		shows[id] = cleaned
	}

	resp := &message.AvailableShows{
		Shows: shows,
	}

	e.server.MessageEditor(server.EncodeServerMessage(resp), name)
	return nil
}

func (e *EditorDriver) onSelectShowActivateShow(name string, _ bool, msg message.ClientMessage) error {
	id := msg.Data.(*message.SelectShow).ShowID

	filepath := e.knownShows[id]

	// show
	_, err := OpenShow(filepath)

	if err != nil {
		log.Printf("Failed to open game in editor: %v", err)

		msg := message.SetEditorError{
			Message: fmt.Sprintf("Failed to open game in the editor: %v", err),
			Code:    1400,
		}

		e.server.MessageEditor(server.EncodeServerMessage(msg), name)
		return errors.New("failed to open game in editor")
	}

	return nil
}

func (e *EditorDriver) onJoinManageEditor(name string, _ bool, _ bool) error {
	//m.sessions[name] = &EditorSession{}
	return nil
}

func (e *EditorDriver) refreshGamesFromDisk() []string {
	files, err := util.GetFilesInDir(DataDir)
	if err != nil {
		// Should never happen since we read these files at startup
		log.Fatalf("Could not enumerate saved files: %v", err)
	}

	e.knownShows = make(map[string]string)
	for _, f := range files {

		hasher := sha512.New()
		hasher.Write([]byte(f))
		sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

		e.knownShows[sha] = f
	}

	return files
}
