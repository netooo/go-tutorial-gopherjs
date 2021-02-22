package models

import "github.com/google/uuid"

type Message struct {
	Author   uuid.UUID
	Nickname string
	Content  string
}
