package game

import (
    "time"
    "log"

    "github.com/baconstrip/kiken/server"
    "github.com/baconstrip/kiken/message"
)

// GameDriver is the main object that manages a game.
type GameDriver struct {
    gameState *GameState
    listenerManager *server.ListenerManager
    server *server.Server
}

func (g *GameDriver) OnJoinSendBoard(name string, host bool) error {
    g.gameState.mu.RLock()
    defer g.gameState.mu.RUnlock()

    b := g.gameState.CurrentBoard()
    if b == nil {
        return nil
    }

    msg := server.EncodeServerMessage(b.Snapshot().ToBoardOverview())

    g.server.MessagePlayer(msg, name)

    return nil
}

func (g *GameDriver) OnSelectQuestionMessageShowQuestion(name string, host bool, msg message.ClientMessage) error {
    g.gameState.mu.Lock()
    defer g.gameState.mu.Unlock()
    if !host {
        return nil
    }

    if g.gameState.currentStatus != STATUS_SHOWING_BOARD {
        return nil
    }

    sel := msg.Data.(*message.SelectQuestion)

    q := g.gameState.FindQuestion(sel.ID)
    if q == nil {
        e := server.EncodeServerMessage(message.ServerError{Error: "Bad question", Code: 2000})
        g.server.MessagePlayer(e, name)
        log.Printf("Bad question from client: %v", sel.ID)
        return nil
    }
    prompt := q.Snapshot().ToQuestionPrompt()
    g.server.MessageAll(server.EncodeServerMessage(prompt))
    g.gameState.currentStatus = STATUS_PRESENTING_QUESTION
    return nil
}

func NewGameDriver(s *server.Server, gs *GameState, lm *server.ListenerManager) *GameDriver {
    g := &GameDriver{
        server: s,
        gameState: gs,
        listenerManager: lm,
    }

    lm.RegisterJoin(g.OnJoinSendBoard)
    lm.RegisterMessage("SelectQuestion", g.OnSelectQuestionMessageShowQuestion)
    return g
}

// Run begins the managing of a game. Runs forever, intended to be called in a
// goroutine.
func (g *GameDriver) Run() {
    g.gameState.mu.Lock()
    g.gameState.currentRound = ICHIBAN
    g.gameState.currentStatus = STATUS_SHOWING_BOARD
    g.gameState.mu.Unlock()
    for {
        time.Sleep(10 * time.Millisecond)
    }
}
