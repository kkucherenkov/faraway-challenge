package transport

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	CloseConnection = iota
	GetChallenge
	Challenge
	GetResource
	Resource
)

const separator = "|"

type Message struct {
	Type int    //type of message
	Data string //payload, could be json, quote or empty
}

func (m *Message) ToString() string {
	return fmt.Sprintf("%d%s%s", m.Type, separator, m.Data)
}

func ParseMessage(str string) (*Message, error) {
	var msgType int

	str = strings.TrimSpace(str)
	parts := strings.Split(str, separator)

	if len(parts) < 1 || len(parts) > 2 {
		return nil, fmt.Errorf("message malformed")
	}

	msgType, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("can't recognize the type of message")
	}

	msg := Message{
		Type: msgType,
	}

	if len(parts) == 2 {
		msg.Data = parts[1]
	}

	return &msg, nil
}
