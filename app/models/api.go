package models

import (
	"encoding/json"
	"github.com/google/uuid"
)

type NewRoomRes struct {
	RoomID uuid.UUID `json:"roomId"`
}

type Event struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (e *Event) Unmarshal(v interface{}) error {
	return json.Unmarshal(e.Data, v)
}
