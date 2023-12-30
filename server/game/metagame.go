package game

import (
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/baconstrip/kiken/common"
	"github.com/baconstrip/kiken/message"
	"github.com/baconstrip/kiken/question"
	"github.com/baconstrip/kiken/server"
)

type MetaGameDriver struct {
	mu *sync.RWMutex

	globalLm *server.ListenerManager
	gameLm   *server.ListenerManager

	config Configuration

	server *server.Server

	gameDriver *GameDriver

	questions []*question.Question

	players map[string]*PlayerStats
	host    *PlayerStats
}

func NewMetaGameDriver(questions []*question.Question, s *server.Server, gameLm *server.ListenerManager, globalLm *server.ListenerManager) *MetaGameDriver {
	config := Configuration{
		ChanceTime:         1 * time.Second,
		DisambiguationTime: 200 * time.Millisecond,
		AnswerTime:         10 * time.Second,
	}

	return &MetaGameDriver{
		gameLm:    gameLm,
		globalLm:  globalLm,
		config:    config,
		server:    s,
		questions: questions,
		players:   make(map[string]*PlayerStats),

		mu: &sync.RWMutex{},
	}
}

func (m *MetaGameDriver) Start() {
	m.globalLm.RegisterMessage("CancelGame", m.onCancelGameCancel)
	m.globalLm.RegisterMessage("StartGame", m.onStartGameStart)
	m.globalLm.RegisterJoin(m.onJoinSendUpdatePlayersAndAddPlayer)
	m.globalLm.RegisterLeave(m.onLeaveMarkDisconnected)
}

// generateUpdatePlayers creates the UpdatePlayers message from the players
// that the game knows about.
// Callers must obtain a mutex before calling.
func (m *MetaGameDriver) generateUpdatePlayers() *message.UpdatePlayers {
	plys := make(map[string]message.Player)
	for name, stats := range m.players {
		plys[name] = message.Player{
			Name:      name,
			Money:     stats.Money,
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
func (m *MetaGameDriver) sendUpdatePlayers() {
	msg := server.EncodeServerMessage(m.generateUpdatePlayers())
	m.server.MessageAll(msg)
}

// ----- Metagame listeners -----

func (m *MetaGameDriver) onJoinSendUpdatePlayersAndAddPlayer(name string, host bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If a player is returning.
	if _, ok := m.players[name]; ok {
		m.players[name].Connected = true
		m.sendUpdatePlayers()
		msg := server.EncodeServerMessage(&message.HostAdd{Name: m.host.Name})
		m.server.MessagePlayer(msg, name)
		return nil
	}

	if m.host != nil && m.host.Name == name {
		m.host.Connected = true
		m.sendUpdatePlayers()
		msg := server.EncodeServerMessage(&message.HostAdd{Name: name})
		m.server.MessageAll(msg)
		return nil
	}

	if host {
		m.host = &PlayerStats{Money: 0, Name: name, Connected: true}
		m.sendUpdatePlayers()
		msg := server.EncodeServerMessage(&message.HostAdd{Name: name})
		m.server.MessageAll(msg)
		return nil
	}

	m.players[name] = &PlayerStats{Money: 0, Name: name, Connected: true}
	if m.host != nil {
		msg := server.EncodeServerMessage(&message.HostAdd{Name: m.host.Name})
		m.server.MessagePlayer(msg, name)
	}

	m.sendUpdatePlayers()

	return nil
}

func (m *MetaGameDriver) onLeaveMarkDisconnected(name string, host bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if host {
		m.host.Connected = false
		return nil
	}

	if m.players[name] == nil {
		return nil
	}

	m.players[name].Connected = false
	m.players[name].Selecting = false

	m.sendUpdatePlayers()

	return nil
}

func (m *MetaGameDriver) onStartGameStart(name string, host bool, _ message.ClientMessage) error {
	g := makeTestGame(m.questions)
	driver := NewGameDriver(m.server, g, m.gameLm, m.config, m)

	m.gameDriver = driver

	driver.StartGame(name)

	return nil
}

func (m *MetaGameDriver) onCancelGameCancel(_ string, host bool, _ message.ClientMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !host {
		return nil
	}

	if m.gameDriver != nil {
		m.gameDriver.EndGame()
	}

	m.gameDriver = nil

	return nil
}

// ------------- testing game helper --------------

func makeTestGame(questions []*question.Question) *Game {
	standardCategories, err := question.CollateFullCategories(questions)
	if err != nil {
		log.Printf("Failed to create categories from questions: %v", err)
	}
	owariCategories := question.CollateLoneQuestions(questions, common.OWARI)
	tiebreakerCategories := question.CollateLoneQuestions(questions, common.TIEBREAKER)

	log.Printf("Loaded %v standard categories, %v Owari, %v Tiebreaker.", len(standardCategories), len(owariCategories), len(tiebreakerCategories))

	// For testing, create a board of the first 4 categories from daiichi/daini,
	// and a question from owari.
	daiichiCount, dainiCount := -1, 0
	var daiichiCats, dainiCats []*question.Category

	for _, c := range standardCategories {
		if daiichiCount == 6 && dainiCount == 6 {
			break
		}

		if daiichiCount < 6 && c.Round == common.DAIICHI {
			daiichiCats = append(daiichiCats, c)
			daiichiCount++
		}
		if dainiCount < 6 && c.Round == common.DAINI {
			dainiCats = append(dainiCats, c)
			dainiCount++
		}
	}

	daiichiBoard := NewBoard(common.DAIICHI, daiichiCats...)
	dainiBoard := NewBoard(common.DAINI, dainiCats...)
	rand.Seed(time.Now().Unix())
	owariBoard := NewBoard(common.OWARI, owariCategories[rand.Intn(len(owariCategories))])

	g := New(daiichiBoard, dainiBoard, owariBoard)
	return g
}
