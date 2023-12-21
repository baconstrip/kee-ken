package question

import (
	"github.com/baconstrip/kiken/common"
	"github.com/baconstrip/kiken/message"
)

type CategoryStateSnapshot struct {
	Name      string
	Round     common.Round
	Questions []QuestionStateSnapshot
}

type QuestionStateSnapshot struct {
	Category string
	Value    int
	Question string
	Answer   string
	Round    common.Round
	Showing  int
	Played   bool

	ID string
}

func (c *Category) State() *CategoryState {
	var qstates []*QuestionState
	for _, q := range c.Questions {
		qstates = append(qstates, q.State())
	}
	return &CategoryState{
		data:      c,
		Questions: qstates,
	}
}

// CategoryState represents a category with stateful data. Member "data" should
// never be modified.
type CategoryState struct {
	data *Category

	Questions []*QuestionState
}

func (c *CategoryState) Snapshot() *CategoryStateSnapshot {
	var qsnaps []QuestionStateSnapshot
	for _, q := range c.Questions {
		snap := q.Snapshot()
		qsnaps = append(qsnaps, *snap)
	}

	return &CategoryStateSnapshot{
		Name:      c.data.Name,
		Round:     c.data.Round,
		Questions: qsnaps,
	}
}

func (q *Question) State() *QuestionState {
	return &QuestionState{
		Data:   q,
		Played: false,
	}
}

// QuestionState represents a question with stateful data. Member "data" should
// never be modified.
type QuestionState struct {
	Data *Question

	Played bool
}

func (q *QuestionState) Snapshot() *QuestionStateSnapshot {
	return &QuestionStateSnapshot{
		Category: q.Data.Category,
		Value:    q.Data.Value,
		Question: q.Data.Question,
		Answer:   q.Data.Answer,
		Round:    q.Data.Round,
		Showing:  q.Data.Showing,
		Played:   q.Played,

		ID: q.Data.ID,
	}
}

func (q *QuestionStateSnapshot) ToQuestionPrompt(includeAnswer bool) *message.QuestionPrompt {
	if includeAnswer {
		return &message.QuestionPrompt{
			Question: q.Question,
			Value:    q.Value,
			Answer:   q.Answer,
			ID:       q.ID,
		}
	}

	return &message.QuestionPrompt{
		Question: q.Question,
		Value:    q.Value,
		ID:       q.ID,
	}
}

func (q *QuestionStateSnapshot) ToQuestionHidden() *message.QuestionHidden {
	return &message.QuestionHidden{
		Value:  q.Value,
		Played: q.Played,
		ID:     q.ID,
	}
}

func (c *CategoryStateSnapshot) ToCategoryOverview() *message.CategoryOverview {
	var questions []*message.QuestionHidden
	for _, q := range c.Questions {
		questions = append(questions, q.ToQuestionHidden())
	}
	return &message.CategoryOverview{
		Name:      c.Name,
		Questions: questions,
	}
}
