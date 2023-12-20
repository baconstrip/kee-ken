package game

import (
	"log"
	"math/rand"
	"time"

	"github.com/baconstrip/kiken/message"
	"github.com/baconstrip/kiken/server"
)

type MetaGameDriver struct {
	globalLm *server.ListenerManager
	gameLm   *server.ListenerManager

	config Configuration

	server *server.Server

	gameDriver *GameDriver

	questions []*Question
}

func NewMetaGameDriver(questions []*Question, s *server.Server, gameLm *server.ListenerManager, globalLm *server.ListenerManager) *MetaGameDriver {
	config := Configuration{
		ChanceTime:         5 * time.Second,
		DisambiguationTime: 200 * time.Millisecond,
		AnswerTime:         10 * time.Second,
	}

	return &MetaGameDriver{
		gameLm:    gameLm,
		globalLm:  globalLm,
		config:    config,
		server:    s,
		questions: questions,
	}
}

func makeTestGame(questions []*Question) *Game {
	standardCategories, err := CollateFullCategories(questions)
	if err != nil {
		log.Printf("Failed to create categories from questions: %v", err)
	}
	owariCategories := CollateLoneQuestions(questions, OWARI)
	tiebreakerCategories := CollateLoneQuestions(questions, TIEBREAKER)

	log.Printf("Loaded %v standard categories, %v Owari, %v Tiebreaker.", len(standardCategories), len(owariCategories), len(tiebreakerCategories))

	// For testing, create a board of the first 4 categories from daiichi/daini,
	// and a question from owari.
	daiichiCount, dainiCount := -1, 0
	var daiichiCats, dainiCats []*Category

	for _, c := range standardCategories {
		if daiichiCount == 4 && dainiCount == 5 {
			break
		}

		if daiichiCount < 4 && c.Round == DAIICHI {
			daiichiCats = append(daiichiCats, c)
			daiichiCount++
		}
		if dainiCount < 4 && c.Round == DAINI {
			dainiCats = append(dainiCats, c)
			dainiCount++
		}
	}

	daiichiBoard := NewBoard(DAIICHI, daiichiCats...)
	dainiBoard := NewBoard(DAINI, dainiCats...)
	rand.Seed(time.Now().Unix())
	owariBoard := NewBoard(OWARI, owariCategories[rand.Intn(len(owariCategories))])

	g := New(daiichiBoard, dainiBoard, owariBoard)
	return g
}

func (m *MetaGameDriver) cancelGame(_ string, host bool, _ message.ClientMessage) error {
	if !host {
		return nil
	}

	if m.gameDriver != nil {
		m.gameDriver.EndGame()
	}

	return nil
}

func (m *MetaGameDriver) Start() {
	m.globalLm.RegisterMessage("CancelGame", m.cancelGame)

	g := makeTestGame(m.questions)
	driver := NewGameDriver(m.server, g, m.gameLm, m.config)

	m.gameDriver = driver
}
