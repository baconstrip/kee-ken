package server

import (
    "fmt"
    "log"
    "reflect"
    "time"
    "bytes"
    "strconv"
    "net/http"
    "html/template"
    "path/filepath"
    "encoding/json"

    "golang.org/x/net/websocket"
    "github.com/baconstrip/kiken/message"
)

const (
    sessionName = "_SESSION"
    maxUint64 uint64 = 18446744073709551615
)

type Server struct {
    index *template.Template
    clientPage *template.Template
    hostPage *template.Template

    sessionManager SessionManager

    listenerManager *ListenerManager

    mux *http.ServeMux
    port int
    passcode string
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
        Code: code,
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
        Name: name,
        ServerPasscode: serverPasscode,
        Passcode: passcode,
        Host: host,
    }

    if authInfo.Name == "" {
        writeError(w, "Name required but not provided", 1013)
        return
    }

    if s.sessionManager.userExists(authInfo.Name, true) {
        writeError(w, "That name is already in use, please choose another", 1015)
        return
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
        writeJSON(w, &message.AuthSuccess{"Successfully joined as host"})
        return
    }

    if authInfo.Passcode == "" {
        writeError(w, "Passcode is required and was not provided.", 1011)
        return
    }

    err = s.sessionManager.createSession(SessionVar{
        name: authInfo.Name,
        passcode: authInfo.Passcode,
        host: false,
    }, w)
    writeJSON(w, &message.AuthSuccess{"Successfully joined as player"})
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
    }
    if err != nil {
        return message.ClientMessage{}, fmt.Errorf("error parsing client message: %v", err)
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
func New(templatePath, staticPath, passcode string, port int, lm *ListenerManager) *Server {
    server := &Server{
        port: port,
        passcode: passcode,
        sessionManager: SessionManager{
            sessions: make(map[SessionID]SessionVar),
            connections: make(map[SessionID]*Connection),
            names: make(map[string]SessionID),
            recentlyDropped: make(map[SessionID]time.Time),
        },
        listenerManager: lm,
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
    return http.ListenAndServe(":" + strconv.Itoa(s.port), s.mux)
}
