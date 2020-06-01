package game

import (
    "strconv"
    "github.com/baconstrip/kiken/message"
)

func (q *QuestionStateSnapshot) ToQuestionHidden() *message.QuestionHidden {
    return &message.QuestionHidden{
        Value: q.Value,
        Played: q.Played,
        ID: q.ID,
    }
}

func (c *CategoryStateSnapshot) ToCategoryOverview() *message.CategoryOverview {
    var questions []*message.QuestionHidden
    for _, q := range c.Questions {
        questions = append(questions, q.ToQuestionHidden())
    }
    return &message.CategoryOverview{
        Name: c.Name,
        Questions: questions,
    }
}

func (b *BoardStateSnapshot) ToBoardOverview() *message.BoardOverview {
    var categories []*message.CategoryOverview
    for _, c := range b.Categories {
        categories = append(categories, c.ToCategoryOverview())
    }
    return &message.BoardOverview{
        Round: strconv.Itoa(int(b.Round)),
        Categories: categories,
    }
}

func (q *QuestionStateSnapshot) ToQuestionPrompt() *message.QuestionPrompt {
    return &message.QuestionPrompt{
        Question: q.Question,
        Value: q.Value,
        ID: q.ID,
    }
}
