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

	sessionManager SessionManager

	globalListenerManager *ListenerManager
	gameListenerManager   *ListenerManager

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
	vars, ok := s.sessionManager.sessions[SessionID(session)]
	s.sessionManager.mu.RUnlock()
	if !ok {
		return 0, SessionVar{}, fmt.Errorf("user not authenticated")
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
						if firstLeave {
							s.globalListenerManager.dispatchLeave(name, host)
							s.gameListenerManager.dispatchLeave(name, host)
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
			if firstLeave {
				s.globalListenerManager.dispatchLeave(name, host)
				s.gameListenerManager.dispatchLeave(name, host)
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
			s.globalListenerManager.dispatchMessage(vars.name, vars.host, msg)
			s.gameListenerManager.dispatchMessage(vars.name, vars.host, msg)
		}
	}
}

func (s *Server) playerInteractiveHandler(ws *websocket.Conn) {
	sid, vars, err := s.verifyAuthenticated(ws.Request())
	if err != nil {
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

	host, err := strconv.ParseBool(hostRaw)
	if err != nil {
		writeError(w, "Bad request", 1003)
		return
	}

	authInfo := &message.AuthInfo{
		Name:           name,
		ServerPasscode: serverPasscode,
		Passcode:       passcode,
		Host:           host,
	}

	if authInfo.Name == "" {
		writeError(w, "Name required but not provided", 1013)
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
		value = &message.NextRound{}
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

// New creates the server that handles all of the communication for the game,
// including serving the pages for logging in, static content, and hosting the
// websocket gameplay. The server processes messages and passes them to other
// parts of the program as messages. As such, a ListenerManager is provided by
// reference from the other parts of the program, to allow other aspects to
// register event listeners.
func New(templatePath, staticPath, passcode string, port int, globalLm *ListenerManager, gameLm *ListenerManager) *Server {
	server := &Server{
		port:     port,
		passcode: passcode,
		sessionManager: SessionManager{
			sessions:        make(map[SessionID]SessionVar),
			connections:     make(map[SessionID]*Connection),
			names:           make(map[string]SessionID),
			recentlyDropped: make(map[SessionID]time.Time),
		},
		globalListenerManager: globalLm,
		gameListenerManager:   gameLm,
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

	server.mux = http.NewServeMux()
	server.mux.HandleFunc("/", server.indexHandler)
	server.mux.HandleFunc("/client", server.clientHandler)
	server.mux.HandleFunc("/host", server.hostHandler)
	server.mux.HandleFunc("/auth", server.authHandler)
	server.mux.Handle("/player_game", websocket.Handler(server.playerInteractiveHandler))
	server.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))
	return server
}

func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(":"+strconv.Itoa(s.port), s.mux)
}
