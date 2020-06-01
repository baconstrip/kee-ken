package game

import (
    "time"

    "github.com/baconstrip/kiken/server"
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
   return  nil
}

func NewGameDriver(s *server.Server, gs *GameState, lm *server.ListenerManager) *GameDriver {
    g := &GameDriver{
        server: s,
        gameState: gs,
        listenerManager: lm,
    }

    lm.RegisterJoin(g.OnJoinSendBoard)
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
