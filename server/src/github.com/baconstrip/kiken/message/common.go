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

// BoardOverview defines a message that the server sends to clients
// representing the board without exposing question information.
type BoardOverview struct {
    Round string
    Categories []*CategoryOverview
}

// CategoryOverview defines a message that the server sends to clients
// representing a category without exposing question information.
type CategoryOverview struct {
    Name string
    Questions []*QuestionHidden
}

// QuestionHidden defines a message that the server sends to clients
// representing a question without exposing the prompt or answer.
type QuestionHidden struct {
    Value int
    Played bool
}

type QuestionPrompt struct {
    Question string
    Value int
}

// PlayerAdded is a message sent by the server when a player joins the game
// to instruct the client to add this player to the UI.
type PlayerAdded struct {
    Name string
    Money int
}

// AuthInfo defines a message that the client sends to the server to provide
// authentication information.
type AuthInfo struct {
    Name string
    ServerPasscode string
    Passcode string
    Host bool
}

type AuthSuccess struct {
    Msg string
}

type ServerError struct {
    Error string
    Code int
}

type ClientTestMessage struct {
    Example string
    Number int
}
