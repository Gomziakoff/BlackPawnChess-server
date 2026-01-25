package gameservice

import "encoding/json"

type IncomingMessage struct {
	T string          `json:"t"`
	D json.RawMessage `json:"d,omitempty"`
}

type OutgoingMessage struct {
	T string      `json:"t"`
	D interface{} `json:"d,omitempty"`
}
