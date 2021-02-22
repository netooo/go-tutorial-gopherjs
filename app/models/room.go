package models

import "github.com/google/uuid"

type Room struct {
	UUID    uuid.UUID
	Members []User
	History []Message
}
