package game

import "sync"

type Status int

const (
    STATUS_UNKNOWN Status = iota
    STATUS_PRESTART
    STATUS_PREPARING
    STATUS_SHOWING_BOARD
)

type Round int

const (
    UNKNOWN Round = iota
    ICHIBAN
    NIBAN
    OWARI
    TIEBREAKER
)

// CreateState is called once a Game struct is completed, and attaches state
// data to the Game, allowing it to be played. 
func (g *Game) CreateState() *GameState {
    var bstates []*BoardState
    for _, b := range g.Boards {
        bstates = append(bstates, b.state())
    }
    return &GameState{
        data: g,
        Boards: bstates,
        currentRound: UNKNOWN,
        currentStatus: STATUS_PRESTART,
    }
}

// GameState represents a game in progress with stateful data. Member "data"
// should never be modified. mu should be locked before changing anything in
// the game's state.
type GameState struct {
    mu sync.RWMutex
    data *Game

    Boards []*BoardState

    currentRound Round
    currentStatus Status
}

func (g *GameState) Snapshot() *GameStateSnapshot {
    var bsnaps []BoardStateSnapshot
    for _, b := range g.Boards {
        snap := b.Snapshot()
        bsnaps = append(bsnaps, *snap)
    }
    return &GameStateSnapshot{
        CurrentRound: g.currentRound,
        CurrentStatus: g.currentStatus,
        Boards: bsnaps,
    }
}

func (g *GameState) CurrentBoard() *BoardState {
    if g.currentRound == UNKNOWN {
        return nil
    }
    return g.Boards[int(g.currentRound) - 1]
}

func (b *Board) state() *BoardState {
    var cstates []*CategoryState
    for _, c := range b.Categories {
        cstates = append(cstates, c.state())
    }
    return &BoardState {
        data: b,
        Categories: cstates,
    }
}

// BoardState represents a board with stateful data. Member "data" should
// never be modified.
type BoardState struct {
    data *Board

    Categories []*CategoryState
}

func (b *BoardState) Snapshot() *BoardStateSnapshot {
    var csnaps []CategoryStateSnapshot
    for _, c := range b.Categories {
        snap := c.Snapshot()
        csnaps = append(csnaps, *snap)
    }
    return &BoardStateSnapshot{
        Categories: csnaps,
        Round: b.data.Round,
    }
}

func (c *Category) state() *CategoryState {
    var qstates []*QuestionState
    for _, q := range c.Questions {
        qstates = append(qstates, q.state())
    }
    return &CategoryState{
        data: c,
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
    for _,  q := range c.Questions {
        snap := q.Snapshot()
        qsnaps = append(qsnaps, *snap)
    }

    return &CategoryStateSnapshot{
        Name: c.data.Name,
        Round: c.data.Round,
        Questions: qsnaps,
    }
}

func (q *Question) state() *QuestionState {
    return &QuestionState{
        data: q,
        Played: false,
    }
}


// QuestionState represents a question with stateful data. Member "data" should
// never be modified.
type QuestionState struct {
    data *Question

    Played bool
}

func (q *QuestionState) Snapshot() *QuestionStateSnapshot {
    return &QuestionStateSnapshot{
        Category: q.data.Category,
        Value: q.data.Value,
        Question: q.data.Question,
        Answer: q.data.Answer,
        Round: q.data.Round,
        Showing: q.data.Showing,
        Played: q.Played,

        ID: q.data.ID,
    }
}

type GameStateSnapshot struct {
    CurrentRound Round
    CurrentStatus Status

    Boards []BoardStateSnapshot
}

type BoardStateSnapshot struct {
    Categories []CategoryStateSnapshot
    Round Round
    // Snapshot does not include pachi, as they are never needed by the
    // snapshot.
}

type CategoryStateSnapshot struct {
    Name string
    Round Round
    Questions []QuestionStateSnapshot
}

type QuestionStateSnapshot struct {
    Category string
    Value int
    Question string
    Answer string
    Round Round
    Showing int
    Played bool

    ID string
}
