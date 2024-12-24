package editor

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/baconstrip/kiken/common"
	"github.com/baconstrip/kiken/game"
	"github.com/baconstrip/kiken/question"
)

// TODO: This needs to be reconcilled with the Game type
type Show struct {
	filepath string
	name     string
	id       string

	Rounds []*game.Board
}

func (s *Show) Save() error {
	questions := []*question.Question{}
	for _, r := range s.Rounds {
		for _, c := range r.Categories {
			questions = append(questions, c.Questions...)
		}
	}
	bytes, err := json.Marshal(questions)
	if err != nil {
		return fmt.Errorf("failed to save game, error marshaling: %v", err)
	}

	filepath := DataDir + s.filepath

	err = os.WriteFile(filepath, bytes, 0o755)
	if err != nil {
		return fmt.Errorf("failed to save game, error saving: %v", err)
	}
	return nil
}

func OpenShow(inputPath string) (*Show, error) {
	questions, err := question.LoadQuestions(inputPath)
	if err != nil {
		return nil, fmt.Errorf("could not open question file %v: %v", inputPath, err)
	}

	categories, err := question.CollateFullCategories(questions, false)
	if err != nil {
		return nil, fmt.Errorf("could not correlate questions during Show load: %v", err)
	}

	var owari *question.Question
	for _, q := range questions {
		if q.Round == common.OWARI {
			owari = q
		}
	}

	daiichiCount, dainiCount := 0, 0
	var daiichiCats, dainiCats []*question.Category
	for _, c := range categories {
		if daiichiCount == 6 {
			break
		}

		if c.Round == common.DAIICHI {
			daiichiCats = append(daiichiCats, c)
			daiichiCount++
		}
	}

	for _, c := range categories {
		if dainiCount == 6 {
			break
		}

		if c.Round == common.DAINI {
			dainiCats = append(dainiCats, c)
			dainiCount++
		}
	}

	owariCategory, err := game.NewCategory(owari)
	if err != nil {
		return nil, fmt.Errorf("failed to make owari category: %v", err)
	}

	daiichiBoard := game.Board{
		Categories: daiichiCats,
		Round:      common.DAIICHI,
	}
	dainiBoard := game.Board{
		Categories: dainiCats,
		Round:      common.DAINI,
	}

	owariBoard := game.Board{
		Categories: []*question.Category{owariCategory},
		Round:      common.OWARI,
	}

	filename := path.Base(inputPath)
	ext := path.Ext(filename)
	cleaned := filename[:len(filename)-len(ext)]

	hasher := sha512.New()
	hasher.Write([]byte(filename))
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	return &Show{
		filepath: inputPath,
		name:     cleaned,
		id:       sha,
		Rounds: []*game.Board{
			&daiichiBoard,
			&dainiBoard,
			&owariBoard,
		},
	}, nil
}

func NewShow(filepath string) *Show {
	hasher := sha512.New()
	hasher.Write([]byte(filepath))
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return &Show{
		filepath: filepath + ".json",
		id:       sha,
		// TODO allow making this a name
		name: filepath,
	}
}
