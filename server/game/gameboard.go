package game

import (
	"fmt"
	"sort"
)

// Game represents the non-stateful data of a game.
type Game struct {
	Boards []*Board
}

// Board represents the non-stateful data of Board.
type Board struct {
	Categories []*Category
	// Points to questions where players are given an opportunity to wager on
	// answer.
	pachi []*Question

	Round Round
}

func New(boards ...*Board) *Game {
	return &Game{
		Boards: boards,
	}
}

func NewBoard(round Round, categories ...*Category) *Board {
	return &Board{
		Categories: categories,
		Round:      round,
	}
}

func NewCategory(questions ...*Question) (*Category, error) {
	var name string
	for _, q := range questions {
		if name != "" && name != q.Category {
			return nil, fmt.Errorf("questions found from different categories: %v, %v", name, q.Category)
		}
		name = q.Category
	}

	sort.Sort(ByValue(questions))
	retVal := &Category{
		Name: name,
	}

	retVal.Questions = append(retVal.Questions, questions...)

	return retVal, nil
}
