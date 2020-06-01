package server

import (
    "log"
    "strconv"
    "time"
    "fmt"
    "net/http"
    "encoding/json"

    "golang.org/x/net/websocket"
)

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
                        name, firstLeave := s.sessionManager.dropConnection(sid)
                        if firstLeave {
                            s.listenerManager.dispatchLeave(name)
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
            name, firstLeave := s.sessionManager.dropConnection(sid)
            if firstLeave {
                s.listenerManager.dispatchLeave(name)
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
            s.listenerManager.dispatchMessage(vars.name, vars.host, msg)
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

    s.listenerManager.dispatchJoin(vars.name, vars.host)

//    overview := s.ga.Boards[0].Snapshot().ToBoardOverview()
//
//    s.sessionManager.writeMessage(sid, encodeServerMessage(overview))
//
//    playerAdded := &message.PlayerAdded{
//        Name: vars.name,
//        Money: 0,
//    }
//
//    s.sessionManager.messageAll(encodeServerMessage(playerAdded))
//
//    // Wait 4 seconds, then send a question to the client, for testing.
//    time.Sleep(4*time.Second)
//
//    q := s.ga.Boards[0].Snapshot().Categories[0].Questions[0].ToQuestionPrompt()
//    s.sessionManager.writeMessage(sid, encodeServerMessage(q))
//
//    // Wait 4 seconds, then send a question to the client, for testing.
//    time.Sleep(4*time.Second)
//
//    q = s.ga.Boards[0].Snapshot().Categories[0].Questions[1].ToQuestionPrompt()
//    s.sessionManager.writeMessage(sid, encodeServerMessage(q))

    // Wait forever
    for {
        time.Sleep(10*time.Minute)
    }
}


