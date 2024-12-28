package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
	"time"
	"unicode"

	"github.com/baconstrip/kiken/message"
	"golang.org/x/net/websocket"
)

const (
	sessionName        = "_SESSION"
	maxUint64   uint64 = 18446744073709551615
)

type Server struct {
	index      *template.Template
	clientPage *template.Template
	hostPage   *template.Template
	editorPage *template.Template

	sessionManager SessionManager

	globalListenerManager *ListenerManager
	gameListenerManager   *ListenerManager
	editorListenerMangaer *ListenerManager

	mux      *http.ServeMux
	port     int
	passcode string
}

func (s *Server) verifyAuthenticated(r *http.Request) (SessionID, SessionVar, error) {
	sessionCookie, err := r.Cookie(sessionName)
	if err != nil {
		return 0, SessionVar{}, err
	}

	session, err := strconv.ParseUint(sessionCookie.Value, 10, 64)
	if err != nil {
		return 0, SessionVar{}, err
	}

	s.sessionManager.mu.RLock()
	defer s.sessionManager.mu.RUnlock()
	vars, ok := s.sessionManager.sessions[SessionID(session)]
	if !ok {
		vars, ok := s.sessionManager.editorSessions[SessionID(session)]
		if !ok {
			return 0, SessionVar{}, fmt.Errorf("user not authenticated")
		}
		return SessionID(session), vars, nil
	}

	return SessionID(session), vars, nil
}

func (s *Server) clientWriter(sid SessionID) {
	for {
		err := s.sessionManager.withConnection(sid, func(c *Connection) error {
			select {
			case msg := <-c.out:
				log.Printf("Sending message to client: %+v", msg.Data)
				out, err := json.Marshal(msg)
				if err != nil {
					log.Printf("Error encoding messsage for client: %v", err)
				}
				if err := websocket.Message.Send(c.soc, string(out)); err != nil {
					go func() {
						name, firstLeave, host := s.sessionManager.dropConnection(sid)

						s.sessionManager.mu.RLock()
						vars, ok := s.sessionManager.sessions[sid]
						if !ok {
							s.sessionManager.mu.RUnlock()
							return
						}
						s.sessionManager.mu.RUnlock()

						if firstLeave {
							if !vars.editor {
								s.globalListenerManager.dispatchLeave(name, host)
								s.gameListenerManager.dispatchLeave(name, host)
							} else {
								s.editorListenerMangaer.dispatchLeave(name, false)
							}
						}
					}()
					log.Printf("Dropping connection to client with session %v because of error sending message: %v", sid, err)
					return err
				}
			default:
			}
			return nil
		})
		if err != nil {
			log.Printf("Dropped connection to %v", sid)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (s *Server) clientReader(sid SessionID, ws *websocket.Conn) {
	for {
		var msg []byte
		err := websocket.Message.Receive(ws, &msg)
		if err != nil {
			log.Printf("Dropping connection to client with session %v because of error reading input: %v", sid, err)
			name, firstLeave, host := s.sessionManager.dropConnection(sid)

			s.sessionManager.mu.RLock()
			vars, ok := s.sessionManager.sessions[sid]
			if !ok {
				s.sessionManager.mu.RUnlock()
				return
			}
			s.sessionManager.mu.RUnlock()

			if firstLeave {
				if !vars.editor {
					s.globalListenerManager.dispatchLeave(name, host)
					s.gameListenerManager.dispatchLeave(name, host)
				} else {
					s.editorListenerMangaer.dispatchLeave(name, false)
				}
			}
			return
		}

		m, err := decodeClientMessage(msg)
		if err != nil {
			log.Printf("Bad message from client %v, error: %v", sid, err)
			continue
		}
		err = s.sessionManager.withConnection(sid, func(c *Connection) error {
			c.in <- m
			return nil
		})
		// This can only error if the connection isn't found, which means it's
		// already been deleted.
		if err != nil {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}
}

func (s *Server) clientDispatcher(sid SessionID) {
	for {
		// Obtain the name and a fresh reference to the input channel.
		s.sessionManager.mu.RLock()

		vars, ok := s.sessionManager.sessions[sid]
		if !ok {
			s.sessionManager.mu.RUnlock()
			return
		}
		conn, ok := s.sessionManager.connections[sid]
		if !ok {
			s.sessionManager.mu.RUnlock()
			return
		}
		inChan := conn.in
		s.sessionManager.mu.RUnlock()

		// Park until a message is available on the input channel, then
		// propagate when a message is found. Alternatively, return if the
		// channel is already closed.
		if msg, ok := <-inChan; !ok {
			log.Printf("leaving client dispatcher, channel is already closed.")
			return
		} else {
			if !vars.editor {
				s.globalListenerManager.dispatchMessage(vars.name, vars.host, msg)
				s.gameListenerManager.dispatchMessage(vars.name, vars.host, msg)
			} else {
				s.editorListenerMangaer.dispatchMessage(vars.name, false, msg)
			}
		}
	}
}

func (s *Server) editorInteractiveHandler(ws *websocket.Conn) {
	sid, vars, err := s.verifyAuthenticated(ws.Request())
	if err != nil {
		ws.Close()
		return
	}

	if !vars.editor {
		log.Printf("Non-editor tried to join as an editor, probable bug: %v", vars.name)
		ws.Close()
		return
	}

	s.sessionManager.addConnection(sid, ws)

	go s.clientWriter(sid)
	go s.clientReader(sid, ws)
	go s.clientDispatcher(sid)

	s.editorListenerMangaer.dispatchJoin(vars.name, false)

	// Wait forever
	for {
		time.Sleep(10 * time.Minute)
	}
}

func (s *Server) playerInteractiveHandler(ws *websocket.Conn) {
	sid, vars, err := s.verifyAuthenticated(ws.Request())
	if err != nil {
		ws.Close()
		return
	}

	if vars.editor {
		// editor trying to join as a player? probably a bug
		log.Printf("Editor attempted to join as a player, probable bug, name: %v", vars.name)
		ws.Close()
		return
	}

	log.Printf("Session of authenticated client: %+v", vars)

	s.sessionManager.addConnection(sid, ws)

	go s.clientWriter(sid)
	go s.clientReader(sid, ws)
	go s.clientDispatcher(sid)

	s.globalListenerManager.dispatchJoin(vars.name, vars.host)
	s.gameListenerManager.dispatchJoin(vars.name, vars.host)

	// Wait forever
	for {
		time.Sleep(10 * time.Minute)
	}
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	err := s.index.ExecuteTemplate(w, "index", nil)
	if err != nil {
		log.Printf("Error loading index page: %v", err)
	}
}

func (s *Server) clientHandler(w http.ResponseWriter, r *http.Request) {
	err := s.clientPage.ExecuteTemplate(w, "client", nil)
	if err != nil {
		log.Printf("Error loading client page: %v", err)
	}
}

func (s *Server) hostHandler(w http.ResponseWriter, r *http.Request) {
	err := s.hostPage.ExecuteTemplate(w, "host", nil)
	if err != nil {
		log.Printf("Error loading client page: %v", err)
	}
}

func (s *Server) editorHandler(w http.ResponseWriter, r *http.Request) {
	err := s.editorPage.ExecuteTemplate(w, "editor", nil)
	if err != nil {
		log.Printf("Error loading editor page: %v", err)
	}
}

func writeError(w http.ResponseWriter, msg string, code int) {
	m, err := json.Marshal(&message.ServerError{
		Error: msg,
		Code:  code,
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	http.Error(w, string(m), 500)
}

func writeJSON(w http.ResponseWriter, msg interface{}) {
	e := json.NewEncoder(w)
	e.Encode(msg)
}

// TODO: make these have their own file
// Check if the string contains no punctuation and only common scripts
func isValidName(str string) bool {
	for _, r := range str {
		// Check for punctuation marks
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			return false
		}

		// Check if the character belongs to a valid script (Latin, Greek, Cyrillic, or CJK)
		if !(unicode.IsLetter(r) || runeIsCJK(r) || runeIsEmoji(r)) {
			return false
		}
	}
	return true
}

// Check if the rune is a CJK character (Chinese, Japanese, Korean)
func runeIsCJK(r rune) bool {
	// Unicode range for CJK characters (Chinese, Japanese, Korean)
	return (r >= 0x4E00 && r <= 0x9FFF) || // CJK Ideographs
		(r >= 0x3040 && r <= 0x309F) || // Hiragana (Japanese)
		(r >= 0x30A0 && r <= 0x30FF) || // Katakana (Japanese)
		(r >= 0xAC00 && r <= 0xD7AF) // Hangul (Korean)
}

// Check if the rune is an emoji character
func runeIsEmoji(r rune) bool {
	// Unicode ranges for emojis
	return (r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
		(r >= 0x1F300 && r <= 0x1F5FF) || // Symbols and pictographs
		(r >= 0x1F680 && r <= 0x1F6FF) || // Transport and map symbols
		(r >= 0x1F700 && r <= 0x1F77F) || // Alchemical symbols
		(r >= 0x2600 && r <= 0x26FF) || // Miscellaneous symbols
		(r >= 0x2700 && r <= 0x27BF) || // Dingbats
		(r >= 0x2B50 && r <= 0x2B50) || // Star emoji
		(r >= 0x1F900 && r <= 0x1F9FF) // Supplemental symbols and pictographs
}

func (s *Server) authHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "Bad request", 1001)
		return
	}

	if err := r.ParseForm(); err != nil {
		writeError(w, "Bad request", 1002)
		return
	}

	// Manually extract the AuthInfo message since it's not sent over the
	// websocket in JSON.
	name := r.PostFormValue("Name")
	serverPasscode := r.PostFormValue("ServerPasscode")
	passcode := r.PostFormValue("Passcode")
	hostRaw := r.PostFormValue("Host")
	editorRaw := r.PostFormValue("Editor")

	host, err := strconv.ParseBool(hostRaw)
	if err != nil {
		writeError(w, "Bad request", 1003)
		return
	}
	editor, err := strconv.ParseBool(editorRaw)
	if err != nil {
		writeError(w, "Bad request", 1090)
		return
	}

	authInfo := &message.AuthInfo{
		Name:           name,
		ServerPasscode: serverPasscode,
		Passcode:       passcode,
		Host:           host,
		Editor:         editor,
	}

	// Various checks
	if len(authInfo.Name) > 50 {
		writeError(w, "Name is too long", 1091)
		return
	}
	if authInfo.Name == "" {
		writeError(w, "Name required but not provided", 1013)
		return
	}

	if !isValidName(authInfo.Name) {
		writeError(w, "Name contains invalid characters", 1092)
		return
	}

	if authInfo.Editor {
		if authInfo.ServerPasscode != s.passcode {
			writeError(w, "Bad passcode", 1008)
			return
		}

		err := s.sessionManager.createSession(SessionVar{
			name:   authInfo.Name,
			editor: true,
		}, w)

		if err != nil {
			writeError(w, "Bad request", 1009)
			log.Printf("Failed to create session for editor: %v", err)
			return
		}

		writeJSON(w, &message.AuthSuccess{Msg: "Successfully joined as editor"})
		return
	}

	if s.sessionManager.userExists(authInfo.Name, true) {
		if !authInfo.Host {
			if s.sessionManager.correctPasscode(authInfo.Name, true, passcode) {
				err = s.sessionManager.createSession(SessionVar{
					name:     authInfo.Name,
					passcode: authInfo.Passcode,
					host:     false,
				}, w)
				if err != nil {
					log.Printf("Failed to create session for returning player: %v", err)
					writeError(w, "Bad request", 1019)
				}

				writeJSON(w, &message.AuthSuccess{Msg: "Successfully rejoined as player"})
				return
			} else {
				writeError(w, "Incorrect passcode", 1015)
				return
			}
		}
	}

	if authInfo.Host {
		if authInfo.ServerPasscode != s.passcode {
			writeError(w, "Bad passcode", 1008)
			return
		}

		err := s.sessionManager.createSession(SessionVar{
			name: authInfo.Name,
			host: true,
		}, w)
		if err != nil {
			writeError(w, "Bad request", 1009)
			log.Printf("Failed to create session for host: %v", err)
			return
		}
		writeJSON(w, &message.AuthSuccess{Msg: "Successfully joined as host"})
		return
	}

	if authInfo.Passcode == "" {
		writeError(w, "Passcode is required and was not provided.", 1011)
		return
	}

	err = s.sessionManager.createSession(SessionVar{
		name:     authInfo.Name,
		passcode: authInfo.Passcode,
		host:     false,
	}, w)

	if err != nil {
		writeError(w, "Bad request", 1012)
		log.Printf("Failed to create session for player: %v", err)
		return
	}

	writeJSON(w, &message.AuthSuccess{Msg: "Successfully joined as player"})
}

func decodeClientMessage(msg []byte) (message.ClientMessage, error) {
	r := bytes.NewReader(msg)
	d := json.NewDecoder(r)

	// Assume there's a starting token.
	_, err := d.Token()
	if err != nil {
		return message.ClientMessage{}, err
	}

	t, err := d.Token()
	if err != nil {
		return message.ClientMessage{}, err
	}
	if s, ok := t.(string); !ok || s != "Type" {
		return message.ClientMessage{}, fmt.Errorf("bad message type to decodeClientMessage, got %v", s)
	}

	// Assume the next value is the type.
	var msgType string
	err = d.Decode(&msgType)

	if err != nil {
		return message.ClientMessage{}, err
	}

	t, err = d.Token()
	if err != nil {
		return message.ClientMessage{}, err
	}
	if s, ok := t.(string); !ok || s != "Data" {
		return message.ClientMessage{}, fmt.Errorf("bad message type to decodeClientMessage, got %v", s)
	}

	var value interface{}
	switch msgType {
	case "ClientTestMessage":
		m := message.ClientTestMessage{}
		err = d.Decode(&m)
		value = &m
	case "SelectQuestion":
		m := message.SelectQuestion{}
		err = d.Decode(&m)
		value = &m
	case "FinishReading":
		value = &message.FinishReading{}
	case "MoveOn":
		value = &message.MoveOn{}
	case "NextRound":
		value = &message.NextRound{}
	case "StartGame":
		value = &message.StartGame{}
	case "AttemptAnswer":
		m := message.AttemptAnswer{}
		err = d.Decode(&m)
		value = &m
	case "MarkAnswer":
		m := message.MarkAnswer{}
		err = d.Decode(&m)
		value = &m
	case "EnterBid":
		m := message.EnterBid{}
		err = d.Decode(&m)
		value = &m
	case "FreeformAnswer":
		m := message.FreeformAnswer{}
		err = d.Decode(&m)
		value = &m
	case "CancelGame":
		m := message.CancelGame{}
		err = d.Decode(&m)
		value = &m

	// Editor messages
	case "RequestShows":
		value = &message.RequestShows{}
	case "SelectShow":
		m := message.SelectShow{}
		err = d.Decode(&m)
		value = &m
	case "AddCategory":
		m := message.AddCategory{}
		err = d.Decode(&m)
		value = &m
	default:
		log.Printf("Unknown message from client: %v", msgType)
		return message.ClientMessage{}, fmt.Errorf("bad message type: %v", msgType)
	}
	if err != nil {
		return message.ClientMessage{}, fmt.Errorf("error parsing client message: %v", err)
	}
	if value == nil {
		return message.ClientMessage{}, fmt.Errorf("nil message from client, discarding")
	}
	log.Printf("Decoded message from client: %+v", value)
	return message.ClientMessage{
		Type: msgType,
		Data: value,
	}, nil
}

// Begin exposed interface.

// EncodeServerMessage wraps a message from the server in message.ServerMessage
// in preparation for sending to a client. msg must be a pointer to the data.
func EncodeServerMessage(msg interface{}) message.ServerMessage {
	name := reflect.TypeOf(msg).Elem().Name()
	return message.ServerMessage{
		Type: name,
		Data: msg,
	}
}

// MessageAll schedules a message to be sent to all clients asynchronously.
// msg should not be modified after calling this function.
func (s *Server) MessageAll(msg message.ServerMessage) {
	s.sessionManager.messageAll(msg)
}

// MessageHost schedules a message to be sent to the host client asynchronously.
// msg should not be modified after calling this function.
func (s *Server) MessageHost(msg message.ServerMessage) {
	s.sessionManager.messageHost(msg)
}

// MessageAll schedules a message to be sent to all player clients
// asynchronously. msg should not be modified after calling this function.
func (s *Server) MessagePlayers(msg message.ServerMessage) {
	s.sessionManager.messagePlayers(msg)
}

// MessagePlayer schedules a message to be sent to the client named by name
// asynchronously. msg should not be modified after calling this function.
func (s *Server) MessagePlayer(msg message.ServerMessage, name string) {
	s.sessionManager.messagePlayer(msg, name)
}

// MessageDitor schedules a message to be sent to the client named by name
// asynchronously. msg should not be modified after calling this function.
func (s *Server) MessageEditor(msg message.ServerMessage, name string) {
	s.sessionManager.messageEditor(msg, name)
}

// New creates the server that handles all of the communication for the game,
// including serving the pages for logging in, static content, and hosting the
// websocket gameplay. The server processes messages and passes them to other
// parts of the program as messages. As such, a ListenerManager is provided by
// reference from the other parts of the program, to allow other aspects to
// register event listeners.
func New(templatePath, staticPath, passcode string, port int, globalLm *ListenerManager, gameLm *ListenerManager, editorLm *ListenerManager) *Server {
	server := &Server{
		port:     port,
		passcode: passcode,
		sessionManager: SessionManager{
			sessions:        make(map[SessionID]SessionVar),
			connections:     make(map[SessionID]*Connection),
			names:           make(map[string]SessionID),
			recentlyDropped: make(map[SessionID]time.Time),
			editorSessions:  make(map[SessionID]SessionVar),
		},
		globalListenerManager: globalLm,
		gameListenerManager:   gameLm,
		editorListenerMangaer: editorLm,
	}
	server.index = template.Must(template.ParseFiles(
		filepath.Join(templatePath, "index.html"),
		filepath.Join(templatePath, "head.html"),
		filepath.Join(templatePath, "nav.html"),
		filepath.Join(templatePath, "jsdefs.html"),
		filepath.Join(templatePath, "pagetop.html"),
	))

	server.clientPage = template.Must(template.ParseFiles(
		filepath.Join(templatePath, "client.html"),
		filepath.Join(templatePath, "head.html"),
		filepath.Join(templatePath, "nav.html"),
		filepath.Join(templatePath, "jsdefs.html"),
		filepath.Join(templatePath, "pagetop.html"),
	))

	server.hostPage = template.Must(template.ParseFiles(
		filepath.Join(templatePath, "host.html"),
		filepath.Join(templatePath, "head.html"),
		filepath.Join(templatePath, "nav.html"),
		filepath.Join(templatePath, "jsdefs.html"),
		filepath.Join(templatePath, "pagetop.html"),
	))

	server.editorPage = template.Must(template.ParseFiles(
		filepath.Join(templatePath, "editor.html"),
		filepath.Join(templatePath, "head.html"),
		filepath.Join(templatePath, "nav.html"),
		filepath.Join(templatePath, "jsdefs.html"),
		filepath.Join(templatePath, "pagetop.html"),
	))

	server.mux = http.NewServeMux()
	server.mux.HandleFunc("/", server.indexHandler)
	server.mux.HandleFunc("/client", server.clientHandler)
	server.mux.HandleFunc("/editor", server.editorHandler)
	server.mux.HandleFunc("/host", server.hostHandler)
	server.mux.HandleFunc("/auth", server.authHandler)
	server.mux.Handle("/player_game", websocket.Handler(server.playerInteractiveHandler))
	server.mux.Handle("/editor_ws", websocket.Handler(server.editorInteractiveHandler))
	server.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))
	return server
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(":"+strconv.Itoa(s.port), s.mux)
}
