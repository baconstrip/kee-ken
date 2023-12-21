package game

import (
	"fmt"
	"sort"

	"github.com/baconstrip/kiken/common"
	"github.com/baconstrip/kiken/question"
)

// Game represents the non-stateful data of a game.
type Game struct {
	Boards []*Board
}

// Board represents the non-stateful data of Board.
type Board struct {
	Categories []*question.Category
	// Points to questions where players are given an opportunity to wager on
	// answer.
	pachi []*question.Question

	Round common.Round
}

func New(boards ...*Board) *Game {
	return &Game{
		Boards: boards,
	}
}

func NewBoard(round common.Round, categories ...*question.Category) *Board {
	return &Board{
		Categories: categories,
		Round:      round,
	}
}

func NewCategory(questions ...*question.Question) (*question.Category, error) {
	var name string
	for _, q := range questions {
		if name != "" && name != q.Category {
			return nil, fmt.Errorf("questions found from different categories: %v, %v", name, q.Category)
		}
		name = q.Category
	}

	sort.Sort(question.ByValue(questions))
	retVal := &question.Category{
		Name: name,
	}

	retVal.Questions = append(retVal.Questions, questions...)

	return retVal, nil
}
