// Package message defines all messages sent between the client and server.
// Types in the this package should rely on primitive types and types in this
// directory only.
package message

// ClientMessage wraps all messages from a client, allowing the server to
// determine the type of the message sent.
type ClientMessage struct {
	// The Type of the message should exactly match one of the messages in this
	// package.
	Type string
	Data interface{}
}

// ServerMessage wraps all message from the server, allowing the client to
// determine the type of the message sent.
type ServerMessage struct {
	Type string
	Data interface{}
}

// ------- BEGIN SERVER MESSAGES --------

// BoardOverview defines a message that the server sends to clients
// representing the board without exposing question information.
type BoardOverview struct {
	Round      string
	Categories []*CategoryOverview
}

// CategoryOverview defines a message that the server sends to clients
// representing a category without exposing question information.
type CategoryOverview struct {
	Name      string
	Questions []*QuestionHidden
}

// QuestionHidden defines a message that the server sends to clients
// representing a question without exposing the prompt or answer.
type QuestionHidden struct {
	Value  int
	Played bool
	ID     string
}

// QuestionPrompt defines a message that the server sends to clients
// to request that question be shown to clients. Answer is only set for the
// host client.
type QuestionPrompt struct {
	Question string
	Value    int
	Answer   string
	ID       string
}

// OpenResponses is a message that the server sends to clients to instruct
// them to begin counting down to give clients a chance to become the one to
// answer a question.
type OpenResponses struct {
	// Interval is the time in milliseconds that players are allowed to request
	// to answer the question.
	Interval int
}

// CloseResponses is a message that the server sends to clients to indicate
// that the server is no longer accepting buzzes.
type CloseResponses struct{}

// HideQuestion is a message that the server sends to clients to indicate
// that the client should no longer show the question prompt.
type HideQuestion struct{}

// PlayerAnswering is a message that the server sends to clients to indicate
// that the player given by Name is attempting to answer the question.
type PlayerAnswering struct {
	Name string
	// Interval is the time in milliseconds that a player is given to answer a
	// question.
	Interval int
}

// UpdatePlayers is a message sent by the server to refresh the players on a
// client.
type UpdatePlayers struct {
	Plys map[string]Player
}

// Player messages are not sent directly, but are embedded in an UpdatePlayers
// message to describe players.
type Player struct {
	Name  string
	Money int

	// Whether the player is currently connected. For transitive disconnects.
	Connected bool
	// Whether the player is currently selecting a question.
	Selecting bool
}

// HostAdd is a message sent by the server when the host joins or to set the
// host for a player that has just joined.
type HostAdd struct {
	Name string
}

// ServerError is sent to the client when the server fails to handle a request.
type ServerError struct {
	Error string
	Code  int
}

type AuthSuccess struct {
	Msg string
}

// BeginOwari is sent to the clients to indicate the beginning of the endgame.
// It contains a single category with a single question.
type BeginOwari struct {
	Category *CategoryOverview
	// The money the client currently has. This is a hack.
	Money int
}

// ShowOwariPrompt is sent to the clients to begin entries for Owari.
type ShowOwariPrompt struct {
	Prompt *QuestionPrompt
}

type ShowOwariResults struct {
	Answers map[string]string
	Bids    map[string]int
}

type ClearBoard struct{}

// ------- EDITOR MESSAGES --------

// AvailableShows is a response to the client's request for shows, and contains
// a map of show IDs to names
type AvailableShows struct {
	Shows map[string]string
}

// SetEditorError shows a client an error in the editor.
type SetEditorError struct {
	Message string
	Code    int
}

type UpdateEditorBoards struct {
}

// ------- BEGIN CLIENT MESSAGES --------

// AuthInfo defines a message that the client sends to the server to provide
// authentication information.
type AuthInfo struct {
	Name           string
	ServerPasscode string
	Passcode       string
	Host           bool
	Editor         bool
	Spectator      bool
}

// SelectQuestion is a message that the clients sends to indicate the question
// with ID should be the next question played.
type SelectQuestion struct {
	ID string
}

// MarkAnswer is a message that the host client sends to decide whether a
// players' answer is correct and to indicate that answering period is over.
type MarkAnswer struct {
	Correct bool
}

// FinishReading is a message that the host client sends when to indicate they
// have finished reading the question, and that the server should start
// accepting requests to answer.
type FinishReading struct{}

// MoveOn is a message that the host client sends to advance the game after a
// quesiton.
type MoveOn struct{}

// NextRound is a message that the client sends to move the game to the next
// round once a round is complete.
type NextRound struct{}

// StartGame is a messsage that the host clients send to start the game from
// a before start state.
type StartGame struct{}

// AttemptAnswer is a message that a player client sends when they want to
// attempt to answer a question.
type AttemptAnswer struct {
	// Amount of time between processing the request to when the client pressed
	// a button requesting to answer, as measured by the client.
	ResponseTime int
}

// EnterBid is a message that a player sends when prompted to wager an amount
// of money.
type EnterBid struct {
	Money int
}

// FreeformAnswer is a message that a player sends to enter a free form text
// message.
type FreeformAnswer struct {
	Message string
}

type ClientTestMessage struct {
	Example string
	Number  int
}

// CancelGame is for the host to indicate the current game should be closed.
type CancelGame struct{}

// ---- EDITOR MESSAGES -----
// Requests that the server show the shows available
type RequestShows struct{}

// Tells the server to select a show for editing
type SelectShow struct {
	ShowID string
}

// Tells the server to add a new category for editing
type AddCategory struct {
	Name  string
	Round string
}

type AdjustScore struct {
	PlayerName string
	Amount     int
}
