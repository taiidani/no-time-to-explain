package models

const MessagesDBKey = "messages"

type Messages struct {
	Messages []Message
}

type Message struct {
	ID       string
	Trigger  string
	Response string
}
