package server

import (
    "log"
    "strconv"
    "net/http"
    "html/template"
    "path/filepath"
    "encoding/json"

    "golang.org/x/net/websocket"
    "github.com/baconstrip/kiken/game"
)

type Server struct {
    index *template.Template
    mux *http.ServeMux
    port int
    passcode string
    ga *game.Game
}

type playerMessage struct {
    Test string
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
    err := s.index.Execute(w, nil)
    if err != nil {
        log.Printf("Error loading index page: %v", err)
    }
}

func (s *Server) playerInteractiveHandler(ws *websocket.Conn) {
    for {
        var response string
        if err := websocket.Message.Receive(ws, &response); err != nil {
            log.Printf("Connection broken: %v", err)
            break
        }

        log.Printf("From the client: %v", response)
        out, err := json.Marshal(s.ga.Boards[0])
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

func New(templatePath, staticPath, passcode string, port int, g *game.Game) *Server {
    server := &Server{
        port: port,
        passcode: passcode,
        ga: g,
    }
    server.index = template.Must(template.ParseFiles(filepath.Join(templatePath, "index.html")))

    server.mux = http.NewServeMux()
    server.mux.HandleFunc("/", server.indexHandler)
    server.mux.Handle("/player_game", websocket.Handler(server.playerInteractiveHandler))
    server.mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))
    return server
}

func (s *Server) ListenAndServe() error {
    return http.ListenAndServe(":" + strconv.Itoa(s.port), s.mux)
}
