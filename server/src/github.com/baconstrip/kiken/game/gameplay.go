package game

import (
    "time"
    "math"
    "fmt"
    "log"
    "math/rand"

    "github.com/baconstrip/kiken/server"
    "github.com/baconstrip/kiken/message"
)


type Configuration struct {
    // How long players are given to press to be able to press the button for a
    // chance to answer the question.
    ChanceTime time.Duration

    // Time to wait after one player buzzes to make sure that that person who
    // buzzed the fastest wins, even if their packet/message gets to the server
    // after someone else's.
    DisambiguationTime time.Duration

    // How long players are given to answer a question.
    AnswerTime time.Duration
}

// GameDriver is the main object that manages a game.
type GameDriver struct {
    // TODO refactor a mutex into this struct, instead of relying on the mutex
    // of the GameState.
    gameState *GameState
    listenerManager *server.ListenerManager
    server *server.Server

    config Configuration

    players map[string]*PlayerStats
    host *PlayerStats
    quesState *questionPromptState
    owariState *owariState
}

type questionPromptState struct {
    question *QuestionState
    questionOpened time.Time
    attemptedBuzzes map[string]int
    alreadyAnswered []string
    buzzCloseTimerSet bool
    playerAnswering string

    buzzTimeoutUnless unlessFunc
}

type owariState struct {
    question *QuestionState
    bids map[string]int
    answers map[string]string
}

type PlayerStats struct {
    Name string
    Money int

    Connected bool
    Selecting bool
}

// generateUpdatePlayers creates the UpdatePlayers message from the players
// that the game knows about.
// Callers must obtain a mutex before calling.
func (g *GameDriver) generateUpdatePlayers() *message.UpdatePlayers {
    plys := make(map[string]message.Player)
    for name, stats := range g.players {
        plys[name] = message.Player{
            Name: name,
            Money: stats.Money,
            Connected: stats.Connected,
            Selecting: stats.Selecting,
        }
    }
    return &message.UpdatePlayers{
        Plys: plys,
    }
}

// sendUpdatePlayers sends an UpdatePlayers message to all clients.
// Callers must obtain a mutex before calling.
func (g *GameDriver) sendUpdatePlayers() {
    msg := server.EncodeServerMessage(g.generateUpdatePlayers())
    g.server.MessageAll(msg)
}

// sendUpdateBoard sends the entire Board to all players, to update the clients
// view of the game board.
// Callers must obtain a mutex before calling.
func (g *GameDriver) sendUpdateBoard() {
    if g.gameState.IsOwariState() {
        g.sendOwari()
        return
    }
    b := g.gameState.CurrentBoard()
    if b == nil {
        return
    }
    overview := server.EncodeServerMessage(b.Snapshot().ToBoardOverview())
    g.server.MessageAll(overview)
}

func (g *GameDriver) sendOwari() {
    cat := g.gameState.Boards[OWARI-1].Categories[0]
    owari := message.BeginOwari{Category: cat.Snapshot().ToCategoryOverview()}
    for _, ply := range g.players {
        plyOwari := owari
        plyOwari.Money = ply.Money
        g.server.MessagePlayer(server.EncodeServerMessage(&plyOwari), ply.Name)
    }
}

// showOwariPrompt should only be called after obtaining the mutex.
func (g *GameDriver) showOwariPrompt() {
    g.gameState.currentStatus = STATUS_OWARI_AWAIT_ANSWERS
    ques := g.gameState.Boards[OWARI-1].Categories[0].Questions[0]
    snap := ques.Snapshot()
    hostPrompt := snap.ToQuestionPrompt(true)
    playerPrompt := snap.ToQuestionPrompt(false)
    g.server.MessagePlayers(server.EncodeServerMessage(&message.ShowOwariPrompt{Prompt: playerPrompt}))
    g.server.MessageHost(server.EncodeServerMessage(&message.ShowOwariPrompt{Prompt: hostPrompt}))
}

// showOwariAnswers should only be called after obtaining the mutex.
func (g *GameDriver) showOwariAnswers() {
    g.gameState.currentStatus = STATUS_SHOWING_OWARI
    msg := message.ShowOwariResults{Answers: g.owariState.answers, Bids: g.owariState.bids}
    g.server.MessageAll(server.EncodeServerMessage(&msg))
}

// playerSelecting returns the Stats struct of the player that is currently
// selecting a question. Returns nil if nobody is selecting.
// Callers must obtain a mutex before calling.
func (g *GameDriver) playerSelecting() *PlayerStats {
    for _, stats := range g.players {
        if stats.Selecting {
            return stats
        }
    }
    return nil
}



// ------ BEGIN LISTENERS -----

func (g *GameDriver) OnJoinSendBoard(name string, host bool) error {
    g.gameState.mu.RLock()
    defer g.gameState.mu.RUnlock()

    if g.gameState.currentStatus == STATUS_PREPARING ||  g.gameState.currentStatus == STATUS_PRESTART {
        return nil
    }

    b := g.gameState.CurrentBoard()
    if b == nil {
        return nil
    }
    msg := server.EncodeServerMessage(b.Snapshot().ToBoardOverview())
    g.server.MessagePlayer(msg, name)

    return nil
}

func (g *GameDriver) OnJoinSendUpdatePlayersAndAddPlayer(name string, host bool) error {
    g.gameState.mu.Lock()
    defer g.gameState.mu.Unlock()

    if g.gameState.currentStatus == STATUS_PLAYERS_ANSWERING || g.gameState.currentStatus == STATUS_PRESENTING_QUESTION || g.gameState.currentStatus == STATUS_PLAYERS_BUZZING {
        snap := g.quesState.question.Snapshot()
        prompt := snap.ToQuestionPrompt(host)
        g.server.MessagePlayer(server.EncodeServerMessage(prompt), name)
     }

    // If a player is returning.
    if _, ok := g.players[name]; ok {
        g.players[name].Connected = true
        g.sendUpdatePlayers()
        msg := server.EncodeServerMessage(&message.HostAdd{Name: g.host.Name})
        g.server.MessagePlayer(msg, name)
        return nil
    }

    if g.host != nil && g.host.Name == name {
        g.host.Connected = true
        g.sendUpdatePlayers()
        msg := server.EncodeServerMessage(&message.HostAdd{Name: name})
        g.server.MessageAll(msg)
        return nil
    }

    if host {
        g.host = &PlayerStats{Money: 0, Name: name, Connected: true}
        g.sendUpdatePlayers()
        msg := server.EncodeServerMessage(&message.HostAdd{Name: name})
        g.server.MessageAll(msg)
        return nil
    }

    g.players[name] = &PlayerStats{Money: 0, Name: name, Connected: true}
    if g.host != nil {
        msg := server.EncodeServerMessage(&message.HostAdd{Name: g.host.Name})
        g.server.MessagePlayer(msg, name)
    }

    g.sendUpdatePlayers()
    return nil
}

func (g *GameDriver) OnLeaveMarkDisconnected(name string, host bool) error {
    g.gameState.mu.Lock()
    defer g.gameState.mu.Unlock()

    if host {
        g.host.Connected = false
        return nil
    }

    if g.players[name] == nil {
        return nil
    }

    g.players[name].Connected = false

    // On the chance they disconnect while answering, reset the state of
    // answering a question.
    if g.gameState.currentStatus == STATUS_PLAYERS_ANSWERING && g.quesState.playerAnswering == name {
        resp := message.OpenResponses{
            Interval: int(g.config.ChanceTime.Seconds() * 1000),
        }
        g.server.MessageAll(server.EncodeServerMessage(&resp))
        g.gameState.currentStatus = STATUS_PLAYERS_BUZZING
        log.Printf("all players counting down")
        g.quesState.questionOpened = time.Now()
        g.quesState.attemptedBuzzes = make(map[string]int)

        unless := runAfterUnless(g.config.ChanceTime, g.TimedTimeOutBuzzing)
        g.quesState.buzzTimeoutUnless = unless
        if s, ok := g.players[g.quesState.playerAnswering]; ok {
            g.players[g.quesState.playerAnswering].Money = s.Money - g.quesState.question.data.Value
        }
        g.sendUpdatePlayers()
    }

    log.Printf("sending updated player lists: %+v", g.players)
    g.sendUpdatePlayers()
    return nil
}

func (g *GameDriver) OnStartGameMessageStartGame(name string, host bool, msg message.ClientMessage) error {
    g.gameState.mu.Lock()
    defer g.gameState.mu.Unlock()

    if g.gameState.currentStatus != STATUS_PRESTART {
        return nil
    }
    if !host {
        return nil
    }

    if len(g.players) == 0 {

        e := server.EncodeServerMessage(&message.ServerError{Error: "Please wait for players to join before starting", Code: 2001})
        g.server.MessagePlayer(e, name)
        return nil
    }

    // Select a random player to begin selecting a question

    idx := rand.Int() % len(g.players)
    i := 0

    // Double randomize, since map iteration order is random.
    for _, stats := range g.players {
        if idx == i {
            stats.Selecting = true
            break
        }
        i++
    }

    g.sendUpdateBoard()
    g.sendUpdatePlayers()

    g.gameState.currentStatus = STATUS_SHOWING_BOARD

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

    log.Printf("Got message from client: %+v", msg)

    sel := msg.Data.(*message.SelectQuestion)

    q := g.gameState.FindQuestion(sel.ID)
    if q == nil {
        e := server.EncodeServerMessage(message.ServerError{Error: "Bad question", Code: 2000})
        g.server.MessagePlayer(e, name)
        log.Printf("Bad question from client: %v", sel.ID)
        return nil
    }
    snap := q.Snapshot()
    playerPrompt := snap.ToQuestionPrompt(false)
    hostPrompt := snap.ToQuestionPrompt(true)
    g.server.MessagePlayers(server.EncodeServerMessage(playerPrompt))
    g.server.MessageHost(server.EncodeServerMessage(hostPrompt))

    g.gameState.currentStatus = STATUS_PRESENTING_QUESTION
    g.quesState = &questionPromptState{
        attemptedBuzzes: make(map[string]int),
        question: q,
    }
    q.Played = true
    return nil
}

func (g *GameDriver) OnFinishReadingMessageBeginCountdown(name string, host bool, msg message.ClientMessage) error {
    g.gameState.mu.Lock()
    defer g.gameState.mu.Unlock()
    if !host {
        return nil
    }

    if g.gameState.currentStatus != STATUS_PRESENTING_QUESTION {
        return nil
    }

    resp := message.OpenResponses{
        Interval: int(g.config.ChanceTime.Seconds() * 1000),
    }
    g.server.MessageAll(server.EncodeServerMessage(&resp))
    g.gameState.currentStatus = STATUS_PLAYERS_BUZZING
    g.quesState.questionOpened = time.Now()

    unless := runAfterUnless(g.config.ChanceTime, g.TimedTimeOutBuzzing)
    g.quesState.buzzTimeoutUnless = unless
    return nil
}

func (g *GameDriver) OnAttemptAnswerMessageAllowAnswer(name string, host bool, msg message.ClientMessage) error {
    g.gameState.mu.Lock()
    defer g.gameState.mu.Unlock()
    if host {
        return nil
    }

    if g.gameState.currentStatus != STATUS_PLAYERS_BUZZING {
        return nil
    }

    time := msg.Data.(*message.AttemptAnswer).ResponseTime
    if dur, ok := g.quesState.attemptedBuzzes[name]; ok {
        if dur < time {
            return nil
        }
    }
    g.quesState.attemptedBuzzes[name] = time
    if g.quesState.buzzCloseTimerSet {
        return nil
    }

    g.quesState.buzzTimeoutUnless()
    runAfter(g.config.DisambiguationTime, g.TimedSelectPlayerToAnswer)

    return nil
}

func (g *GameDriver) OnMarkAnswerMessageMoveAlong(name string, host bool, msg message.ClientMessage) error {
    g.gameState.mu.Lock()
    defer g.gameState.mu.Unlock()

    if !host {
        return nil
    }

    if g.gameState.currentStatus != STATUS_PLAYERS_ANSWERING {
        return nil
    }

    correct := msg.Data.(*message.MarkAnswer).Correct

    if correct {
        g.sendUpdateBoard()
        g.gameState.currentStatus = STATUS_POST_QUESTION

        g.server.MessageAll(server.EncodeServerMessage(&message.CloseResponses{}))
        log.Printf("map: %p", &g.players)
        if s, ok := g.players[g.quesState.playerAnswering]; ok {
            log.Printf("found, granting %v", s)
            g.players[g.quesState.playerAnswering].Money = s.Money + g.quesState.question.data.Value
        }
        // When a player gets a question correct, they get to pick next.
        g.playerSelecting().Selecting = false
        g.players[g.quesState.playerAnswering].Selecting = true

        g.sendUpdatePlayers()
        return nil
    }

    resp := message.OpenResponses{
        Interval: int(g.config.ChanceTime.Seconds() * 1000),
    }
    g.server.MessageAll(server.EncodeServerMessage(&resp))
    g.gameState.currentStatus = STATUS_PLAYERS_BUZZING
    log.Printf("all players counting down")
    g.quesState.questionOpened = time.Now()
    g.quesState.attemptedBuzzes = make(map[string]int)

    unless := runAfterUnless(g.config.ChanceTime, g.TimedTimeOutBuzzing)
    g.quesState.buzzTimeoutUnless = unless
    if s, ok := g.players[g.quesState.playerAnswering]; ok {
        g.players[g.quesState.playerAnswering].Money = s.Money - g.quesState.question.data.Value
    }
    g.quesState.alreadyAnswered = append(g.quesState.alreadyAnswered, g.quesState.playerAnswering)
    g.sendUpdatePlayers()
    return nil
}

func (g *GameDriver) OnMoveOnMessageShowBoard(name string, host bool, msg message.ClientMessage) error {
    g.gameState.mu.Lock()
    defer g.gameState.mu.Unlock()

    if !host {
        return nil
    }

    if g.gameState.currentStatus != STATUS_POST_QUESTION {
        return nil
    }
    g.sendUpdateBoard()

    g.server.MessageAll(server.EncodeServerMessage(&message.HideQuestion{}))
    g.gameState.currentStatus = STATUS_SHOWING_BOARD
    return nil
}

func (g *GameDriver) OnNextRoundMessageAdvanceRound(name string, host bool, msg message.ClientMessage) error {
    g.gameState.mu.Lock()
    defer g.gameState.mu.Unlock()

    if !host {
        return nil
    }

    if g.gameState.currentStatus != STATUS_SHOWING_BOARD {
        return nil
    }

    g.gameState.currentRound = g.gameState.currentRound + 1
    if g.gameState.currentRound == OWARI {
        g.gameState.currentStatus = STATUS_ACCEPTING_BIDS
    }

    lowestValue := math.MaxInt32
    lowestPlayer := ""
    for _, ply := range g.players {
        ply.Selecting = false
        if !ply.Connected {
            continue
        }

        if ply.Money < lowestValue {
            lowestValue = ply.Money
            lowestPlayer = ply.Name
        }
    }

    g.players[lowestPlayer].Selecting = true

    g.sendUpdateBoard()
    g.sendUpdatePlayers()

    return nil
}

func (g *GameDriver) OnEnterBidAddBid(name string, host bool, msg message.ClientMessage) error {
    g.gameState.mu.Lock()
    defer g.gameState.mu.Unlock()

    if host {
        return nil
    }
    bid := msg.Data.(*message.EnterBid).Money

    g.owariState.bids[name] = bid

    // If the bid is less than the amount they have or negative ignore it.
    currentMoney := g.players[name].Money
    if bid > currentMoney || currentMoney < 0 || bid < 0 {
        return nil
    }

    // Check to see if all bids are in.
    found := true
    for n, ply := range g.players {
        if _, ok := g.owariState.bids[n]; !ok {
            // Ignore players with zero or negative money.
            if ply.Money <= 0 {
                continue
            }
            found = false
        }
    }

    if !found {
        return nil
    }

    log.Printf("All bids are in")
    g.showOwariPrompt()

    return nil
}

func (g *GameDriver) OnFreeformAnswerAddAnswerOwari(name string, host bool, msg message.ClientMessage) error {
    g.gameState.mu.Lock()
    defer g.gameState.mu.Unlock()

    if host {
        return nil
    }

    if g.gameState.currentStatus != STATUS_OWARI_AWAIT_ANSWERS {
        return nil
    }

    ans := msg.Data.(*message.FreeformAnswer).Message
    g.owariState.answers[name] = ans
    // Check to see if all answers are in.
    found := true
    for n, _ := range g.players {
        // Ignore players that didn't bid.
        if _, ok := g.owariState.bids[n]; !ok {
            continue
        }
        if _, ok := g.owariState.answers[n]; !ok {
            found = false
        }
    }

    if !found {
        return nil
    }
    log.Printf("All answers are in")

    g.showOwariAnswers()

    return nil
}

func (g *GameDriver) TimedSelectPlayerToAnswer() error {
    g.gameState.mu.Lock()
    defer g.gameState.mu.Unlock()

    var ply string
    lowest := math.MaxInt32
    for p, d := range g.quesState.attemptedBuzzes {
        if d < lowest {
            lowest = d
            ply = p
        }
    }

    if ply == "" {
        return fmt.Errorf("buzz process handler scheduled despite no buzzes")
    }

    g.gameState.currentStatus = STATUS_PLAYERS_ANSWERING
    g.quesState.playerAnswering = ply
    answering := &message.PlayerAnswering{
        Name: ply,
        Interval: int(g.config.AnswerTime.Seconds()*1000),
    }
    g.server.MessageAll(server.EncodeServerMessage(answering))
    return nil
}

func (g *GameDriver) TimedTimeOutBuzzing() error {
    g.gameState.mu.Lock()
    defer g.gameState.mu.Unlock()

    g.gameState.currentStatus = STATUS_POST_QUESTION

    g.server.MessageAll(server.EncodeServerMessage(&message.CloseResponses{}))
    return nil
}

type timedFunc func() error

// Functions scheduled to run after should re-obtain mutexes, as they are run
// asynchronously to the caller.
func runAfter(d time.Duration, f timedFunc) {
    go func(d time.Duration, f timedFunc) {
        select{
        case <-time.After(d):
            err := f()
            if err != nil {
                log.Printf("Error running timed function: %v", err)
            }
        }
    }(d, f)
}

type unlessFunc func()

// runAfterUnless runs a given function after the delay in d, unless returned
// function returned is called, in which case it will do nothing.
func runAfterUnless(d time.Duration, f timedFunc) unlessFunc {
    u := make(chan bool, 100)
    go func(d time.Duration, f timedFunc) {
        select{
        case <-u:
            return
        case <-time.After(d):
            err := f()
            if err != nil {
                log.Printf("Error running timed unless function: %v", err)
            }
        }
    }(d, f)
    return func() {
        u <- true
    }
}

func NewGameDriver(s *server.Server, gs *GameState, lm *server.ListenerManager, config Configuration) *GameDriver {
    g := &GameDriver{
        server: s,
        gameState: gs,
        listenerManager: lm,
        players: make(map[string]*PlayerStats),
        config: config,
        owariState: &owariState{bids: make(map[string]int), answers: make(map[string]string)},
    }

    lm.RegisterJoin(g.OnJoinSendBoard)
    lm.RegisterJoin(g.OnJoinSendUpdatePlayersAndAddPlayer)
    lm.RegisterLeave(g.OnLeaveMarkDisconnected)
    lm.RegisterMessage("StartGame", g.OnStartGameMessageStartGame)
    lm.RegisterMessage("SelectQuestion", g.OnSelectQuestionMessageShowQuestion)
    lm.RegisterMessage("FinishReading", g.OnFinishReadingMessageBeginCountdown)
    lm.RegisterMessage("AttemptAnswer", g.OnAttemptAnswerMessageAllowAnswer)
    lm.RegisterMessage("MarkAnswer", g.OnMarkAnswerMessageMoveAlong)
    lm.RegisterMessage("MoveOn", g.OnMoveOnMessageShowBoard)
    lm.RegisterMessage("NextRound", g.OnNextRoundMessageAdvanceRound)
    lm.RegisterMessage("EnterBid", g.OnEnterBidAddBid)
    lm.RegisterMessage("FreeformAnswer", g.OnFreeformAnswerAddAnswerOwari)

    return g
}

// Run begins the managing of a game. Runs forever, intended to be called in a
// goroutine.
func (g *GameDriver) Run() {
    g.gameState.mu.Lock()
    g.gameState.currentRound = DAIICHI
    g.gameState.currentStatus = STATUS_PRESTART
    g.gameState.mu.Unlock()
    for {
        time.Sleep(10 * time.Millisecond)
    }
}
