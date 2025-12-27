package game

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/baconstrip/kiken/common"
	"github.com/baconstrip/kiken/message"
	"github.com/baconstrip/kiken/question"
	"github.com/baconstrip/kiken/server"
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

	// StartingPhase is the phase the game should start at.
	// one of: "daiichi", "daini", "owari"
	StartingPhase string
}

// GameDriver is the main object that manages a game.
type GameDriver struct {
	// TODO refactor a mutex into this struct, instead of relying on the mutex
	// of the GameState.
	gameState       *GameState
	server          *server.Server
	listenerManager *server.ListenerManager

	config Configuration

	quesState  *questionPromptState
	owariState *owariState

	metagame *MetaGameDriver
}

type questionPromptState struct {
	question          *question.QuestionState
	questionOpened    time.Time
	attemptedBuzzes   map[string]int
	alreadyAnswered   []string
	buzzCloseTimerSet bool
	playerAnswering   string

	buzzTimeoutUnless unlessFunc
}

type owariState struct {
	bids    map[string]int
	answers map[string]string
}

type PlayerStats struct {
	Name  string
	Money int

	Connected bool
	Selecting bool
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

func (g *GameDriver) sendOwariHost() {
	cat := g.gameState.Boards[common.OWARI-1].Categories[0]
	owari := message.BeginOwari{Category: cat.Snapshot().ToCategoryOverview()}
	g.server.MessageHost(server.EncodeServerMessage(&owari))
}

func (g *GameDriver) sendOwariPlayer(name string) {
	cat := g.gameState.Boards[common.OWARI-1].Categories[0]
	owari := message.BeginOwari{Category: cat.Snapshot().ToCategoryOverview()}
	plyOwari := owari
	plyOwari.Money = g.metagame.players[name].Money
	g.server.MessagePlayer(server.EncodeServerMessage(&plyOwari), name)
}

func (g *GameDriver) sendOwari() {
	g.sendOwariHost()
	for name := range g.metagame.players {
		g.sendOwariPlayer(name)
	}

	overview := server.EncodeServerMessage(g.gameState.Boards[common.OWARI-1].Snapshot().ToBoardOverview())
	g.server.MessageAll(overview)
}

func (g *GameDriver) showOwariPromptHost() {
	ques := g.gameState.Boards[common.OWARI-1].Categories[0].Questions[0]
	snap := ques.Snapshot()
	hostPrompt := snap.ToQuestionPrompt(true)
	g.server.MessageHost(server.EncodeServerMessage(&message.ShowOwariPrompt{Prompt: hostPrompt}))
}

func (g *GameDriver) showOwariPromptPlayer(name string) {
	ques := g.gameState.Boards[common.OWARI-1].Categories[0].Questions[0]
	snap := ques.Snapshot()
	playerPrompt := snap.ToQuestionPrompt(false)
	g.server.MessagePlayer(server.EncodeServerMessage(&message.ShowOwariPrompt{Prompt: playerPrompt}), name)
}

// showOwariPrompt should only be called after obtaining the mutex.
func (g *GameDriver) showOwariPrompt() {
	g.gameState.currentStatus = STATUS_OWARI_AWAIT_ANSWERS
	g.showOwariPromptHost()
	for name := range g.metagame.players {
		g.showOwariPromptPlayer(name)
	}
}

func (g *GameDriver) showOwariAnswer(name string) {
	msg := message.ShowOwariResults{Answers: g.owariState.answers, Bids: g.owariState.bids}
	g.server.MessagePlayer(server.EncodeServerMessage(&msg), name)
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
	for _, stats := range g.metagame.players {
		if stats.Selecting {
			return stats
		}
	}
	return nil
}

// ------ BEGIN LISTENERS -----

func (g *GameDriver) OnJoinSendBoard(name string, host bool, spectator bool) error {
	g.gameState.mu.RLock()
	defer g.gameState.mu.RUnlock()

	if g.gameState.currentStatus == STATUS_PREPARING || g.gameState.currentStatus == STATUS_PRESTART {
		return nil
	}

	b := g.gameState.CurrentBoard()
	if b == nil {
		return nil
	}

	if !g.gameState.IsOwariState() {
		msg := server.EncodeServerMessage(b.Snapshot().ToBoardOverview())
		g.server.MessagePlayer(msg, name)
	} else {
		switch g.gameState.currentStatus {
		case STATUS_ACCEPTING_BIDS:
			g.sendOwariPlayer(name)
		case STATUS_OWARI_AWAIT_ANSWERS:
			g.sendOwariPlayer(name)
			g.showOwariPromptPlayer(name)
		case STATUS_SHOWING_OWARI:
			g.sendOwariPlayer(name)
			g.showOwariPromptPlayer(name)
			g.showOwariAnswer(name)
		default:
			log.Fatalf("Unhandled Owari state on join: %v", g.gameState.currentStatus)
		}
	}

	return nil
}

func (g *GameDriver) OnJoinShowQuestionPrompt(name string, host bool, spectator bool) error {
	g.gameState.mu.RLock()
	defer g.gameState.mu.RUnlock()

	if g.gameState.currentStatus == STATUS_PLAYERS_ANSWERING || g.gameState.currentStatus == STATUS_PRESENTING_QUESTION || g.gameState.currentStatus == STATUS_PLAYERS_BUZZING {
		snap := g.quesState.question.Snapshot()
		prompt := snap.ToQuestionPrompt(host)
		g.server.MessagePlayer(server.EncodeServerMessage(prompt), name)
	}

	return nil
}

func (g *GameDriver) makeLowestPlayerSelect() {
	lowestValue := math.MaxInt32
	lowestPlayer := ""
	for _, ply := range g.metagame.players {
		ply.Selecting = false
		if !ply.Connected {
			continue
		}

		if ply.Money < lowestValue {
			lowestValue = ply.Money
			lowestPlayer = ply.Name
		}
	}

	if lowestPlayer != "" {
		g.metagame.players[lowestPlayer].Selecting = true
	}
}

func (g *GameDriver) OnLeaveStopAnswering(name string, host bool, spectator bool) error {
	g.gameState.mu.Lock()
	defer g.gameState.mu.Unlock()

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
		if s, ok := g.metagame.players[g.quesState.playerAnswering]; ok {
			g.metagame.players[g.quesState.playerAnswering].Money = s.Money - g.quesState.question.Data.Value
		}
	}

	foundSelector := false

	for _, ply := range g.metagame.players {
		if ply.Selecting {
			foundSelector = true
			break
		}
	}

	if _, ok := g.metagame.players[name]; ok {
		if g.metagame.players[name].Selecting {
			g.makeLowestPlayerSelect()
		}
	} else if !foundSelector {
		g.makeLowestPlayerSelect()
	}

	g.metagame.sendUpdatePlayers()
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
		e := server.EncodeServerMessage(&message.ServerError{Error: "Bad question", Code: 2000})
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
		question:        q,
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

	// Players who have already tried to answer may not try again.
	for _, n := range g.quesState.alreadyAnswered {
		if n == name {
			return nil
		}
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
		log.Printf("map: %p", &g.metagame.players)
		if s, ok := g.metagame.players[g.quesState.playerAnswering]; ok {
			log.Printf("found, granting %v", s)
			g.metagame.players[g.quesState.playerAnswering].Money = s.Money + g.quesState.question.Data.Value
		}
		// When a player gets a question correct, they get to pick next.
		// fix this shit
		g.playerSelecting().Selecting = false
		g.metagame.players[g.quesState.playerAnswering].Selecting = true

		g.metagame.sendUpdatePlayers()
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
	if s, ok := g.metagame.players[g.quesState.playerAnswering]; ok {
		g.metagame.players[g.quesState.playerAnswering].Money = s.Money - g.quesState.question.Data.Value
	}
	g.quesState.alreadyAnswered = append(g.quesState.alreadyAnswered, g.quesState.playerAnswering)
	g.metagame.sendUpdatePlayers()
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
	if g.gameState.currentRound == common.OWARI {
		g.gameState.currentStatus = STATUS_ACCEPTING_BIDS
	}

	g.makeLowestPlayerSelect()

	g.sendUpdateBoard()
	g.metagame.sendUpdatePlayers()

	return nil
}

func (g *GameDriver) OnEnterBidAddBid(name string, host bool, msg message.ClientMessage) error {
	g.gameState.mu.Lock()
	defer g.gameState.mu.Unlock()

	if host {
		return nil
	}

	if g.gameState.currentStatus != STATUS_ACCEPTING_BIDS {
		return nil
	}

	bid := msg.Data.(*message.EnterBid).Money

	g.owariState.bids[name] = bid

	// If the bid is less than the amount they have or negative ignore it.
	currentMoney := g.metagame.players[name].Money
	if bid > currentMoney || currentMoney < 0 || bid < 0 {
		return nil
	}

	// Check to see if all bids are in.
	found := true
	for n, ply := range g.metagame.players {
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

func (g *GameDriver) OnAdjustScoreMessage(name string, host bool, msg message.ClientMessage) error {
	g.gameState.mu.Lock()
	defer g.gameState.mu.Unlock()

	if !host {
		return nil
	}
	adj := msg.Data.(*message.AdjustScore)

	// Check which player
	if ply, ok := g.metagame.players[adj.PlayerName]; ok {
		ply.Money += adj.Amount
		g.metagame.sendUpdatePlayers()
	} else {
		e := server.EncodeServerMessage(&message.ServerError{Error: "Player not found", Code: 3001})
		g.server.MessagePlayer(e, name)
		log.Printf("Adjust score: player not found: %v", adj.PlayerName)
	}
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
	for n := range g.metagame.players {
		// Ignore players that didn't bid.
		if _, ok := g.owariState.bids[n]; !ok {
			continue
		}
		if _, ok := g.owariState.answers[n]; !ok {
			log.Printf("Still waiting on answer from %v", n)
			found = false
		}
	}

	if !found {
		return nil
	}

	g.gameState.currentStatus = STATUS_SHOWING_OWARI
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
		Name:     ply,
		Interval: int(g.config.AnswerTime.Seconds() * 1000),
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

func (g *GameDriver) EndGame() {
	g.gameState.mu.Lock()
	defer g.gameState.mu.Unlock()

	log.Print("Cancelling game!")

	g.gameState.currentStatus = STATUS_PRESTART

	for _, p := range g.metagame.players {
		p.Money = 0
	}

	g.metagame.sendUpdatePlayers()

	g.sendUpdateBoard()

	g.server.MessageAll(server.EncodeServerMessage(&message.ClearBoard{}))

	g.listenerManager.ClearListeners()
}

type timedFunc func() error

// Functions scheduled to run after should re-obtain mutexes, as they are run
// asynchronously to the caller.
func runAfter(d time.Duration, f timedFunc) {
	go func(d time.Duration, f timedFunc) {
		<-time.After(d)
		err := f()
		if err != nil {
			log.Printf("Error running timed function: %v", err)
		}
	}(d, f)
}

type unlessFunc func()

// runAfterUnless runs a given function after the delay in d, unless returned
// function returned is called, in which case it will do nothing.
func runAfterUnless(d time.Duration, f timedFunc) unlessFunc {
	u := make(chan bool, 100)
	go func(d time.Duration, f timedFunc) {
		select {
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

func NewGameDriver(s *server.Server, game *Game, lm *server.ListenerManager, config Configuration, metagame *MetaGameDriver) *GameDriver {
	gs := game.CreateState()
	driver := &GameDriver{
		server:          s,
		gameState:       gs,
		config:          config,
		listenerManager: lm,
		owariState:      &owariState{bids: make(map[string]int), answers: make(map[string]string)},
		metagame:        metagame,
	}

	lm.RegisterJoin(driver.OnJoinSendBoard)
	lm.RegisterJoin(driver.OnJoinShowQuestionPrompt)
	lm.RegisterLeave(driver.OnLeaveStopAnswering)
	lm.RegisterMessage("SelectQuestion", driver.OnSelectQuestionMessageShowQuestion)
	lm.RegisterMessage("FinishReading", driver.OnFinishReadingMessageBeginCountdown)
	lm.RegisterMessage("AttemptAnswer", driver.OnAttemptAnswerMessageAllowAnswer)
	lm.RegisterMessage("MarkAnswer", driver.OnMarkAnswerMessageMoveAlong)
	lm.RegisterMessage("MoveOn", driver.OnMoveOnMessageShowBoard)
	lm.RegisterMessage("NextRound", driver.OnNextRoundMessageAdvanceRound)
	lm.RegisterMessage("EnterBid", driver.OnEnterBidAddBid)
	lm.RegisterMessage("FreeformAnswer", driver.OnFreeformAnswerAddAnswerOwari)
	lm.RegisterMessage("AdjustScore", driver.OnAdjustScoreMessage)

	switch config.StartingPhase {
	case "daini":
		driver.gameState.currentRound = common.DAINI
	case "owari":
		driver.gameState.currentRound = common.OWARI
		// give all players some starting money to allow for bids
		for _, ply := range metagame.players {
			ply.Money = 1000
		}
	default:
		// Default to daiichi
		driver.gameState.currentRound = common.DAIICHI
	}

	driver.gameState.currentStatus = STATUS_PRESTART

	return driver
}

// StartGames starts the game play, requires the name of the player that
// requested the start.
func (g *GameDriver) StartGame(name string) error {
	g.gameState.mu.Lock()
	defer g.gameState.mu.Unlock()

	if g.gameState.currentStatus != STATUS_PRESTART {
		return nil
	}

	if len(g.metagame.players) == 0 {
		e := server.EncodeServerMessage(&message.ServerError{Error: "Please wait for players to join before starting", Code: 2001})
		g.server.MessagePlayer(e, name)
		return nil
	}

	// Select a random player to begin selecting a question

	idx := rand.Int() % len(g.metagame.players)
	i := 0

	// Double randomize, since map iteration order is random.
	for _, stats := range g.metagame.players {
		if idx == i {
			stats.Selecting = true
			break
		}
		i++
	}

	if g.config.StartingPhase == "owari" {
		g.gameState.currentStatus = STATUS_ACCEPTING_BIDS
	} else {
		g.gameState.currentStatus = STATUS_SHOWING_BOARD
	}

	g.sendUpdateBoard()
	g.metagame.sendUpdatePlayers()

	//g.gameState.currentStatus = STATUS_SHOWING_BOARD

	return nil
}
