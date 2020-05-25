package server

import (
    "log"
    "time"
    "strconv"
    "net/http"
    "html/template"
    "path/filepath"
    "encoding/json"

    "golang.org/x/net/websocket"
)

type Server struct {
    index *template.Template
    mux *http.ServeMux
    port int
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
        message := &playerMessage{Test: "hello client"}
        out, err := json.Marshal(message)
        if err != nil {
            log.Printf("Error marshaling player message: %v", err)
        }
        log.Printf("sending: %v", string(out))
        if err := websocket.Message.Send(ws, string(out)); err != nil{
            log.Printf("Connection broken: %v", err)
            break
        }
        go spam(ws)
    }
}

func spam(ws *websocket.Conn) {
    for {
        time.Sleep(time.Second)
        message := &playerMessage{Test: "hello client"}
        out, err := json.Marshal(message)
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

func New(templatePath string, staticPath string, port int) *Server {
    server := &Server{
        port: port,
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
