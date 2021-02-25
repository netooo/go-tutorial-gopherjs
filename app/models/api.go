package models

import "github.com/google/uuid"

type NewRoomRes struct {
	RoomID uuid.UUID `json:"roomId"`
}
