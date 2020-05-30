package server

import (
    "sync"
    "fmt"
    "time"
    "log"
    "strconv"
    "net/http"
    "html/template"
    "path/filepath"
    "encoding/json"
    "crypto/rand"
    "math/big"

    "golang.org/x/net/websocket"
    "github.com/baconstrip/kiken/game"
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

    mux *http.ServeMux
    port int
    passcode string
    ga *game.GameState
}

type SessionID uint64

type SessionManager struct {
    mu sync.Mutex

    sessions map[SessionID]SessionVar
}

type SessionVar struct {
    name string
    passcode string
    host bool
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

func (s *Server) createSession(vars SessionVar, w http.ResponseWriter) error {
    s.sessionManager.mu.Lock()
    defer s.sessionManager.mu.Unlock()

    var max big.Int
    max.SetUint64(maxUint64)
    keyBig, err := rand.Int(rand.Reader, &max)
    if err != nil {
        return fmt.Errorf("failed to generate random number: %v", err)
    }

    key := keyBig.Uint64()

    s.sessionManager.sessions[SessionID(key)] = vars
    cookie := http.Cookie{
        Name: sessionName,
        Value: strconv.FormatUint(key, 10),
        Expires: time.Now().Add(time.Hour*24),
    }
    http.SetCookie(w, &cookie)
    return nil
}

func (s *Server) destroySession(key SessionID) {
    s.sessionManager.mu.Lock()
    defer s.sessionManager.mu.Unlock()
    delete(s.sessionManager.sessions, key)
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

    if authInfo.Host {
        if authInfo.ServerPasscode != s.passcode {
            writeError(w, "Bad passcode", 1008)
            return
        }

        err := s.createSession(SessionVar{
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

    err = s.createSession(SessionVar{
        name: authInfo.Name,
        passcode: authInfo.Passcode,
        host: false,
    }, w)
    writeJSON(w, &message.AuthSuccess{"Successfully joined as player"})
}

func (s *Server) playerInteractiveHandler(ws *websocket.Conn) {
    r := ws.Request()
    sessionCookie, err := r.Cookie(sessionName)
    if err != nil {
        ws.Close()
        return
    }

    session, err := strconv.ParseUint(sessionCookie.Value, 10, 64)
    if err != nil {
        ws.Close()
        return
    }

    s.sessionManager.mu.Lock()
    vars, ok := s.sessionManager.sessions[SessionID(session)]
    s.sessionManager.mu.Unlock()

    if !ok {
        ws.Close()
        return
    }

    log.Printf("Session of authenticated client: %+v", vars)

    for {
        var response string
        if err := websocket.Message.Receive(ws, &response); err != nil {
            log.Printf("Connection broken: %v", err)
            break
        }

        log.Printf("From the client: %v", response)
        out, err := json.Marshal(s.ga.Boards[0].Snapshot().ToBoardOverview())
        if err != nil {
            log.Printf("Error marshaling player message: %v", err)
        }
        log.Printf("sending: %v", string(out))
        if err := websocket.Message.Send(ws, string(out)); err != nil{
            log.Printf("Connection broken: %v", err)
            break
        }
    }
}

func New(templatePath, staticPath, passcode string, port int, g *game.GameState) *Server {
    server := &Server{
        port: port,
        passcode: passcode,
        ga: g,
        sessionManager: SessionManager{
            sessions: make(map[SessionID]SessionVar),
        },
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
