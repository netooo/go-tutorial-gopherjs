package backend

import (
	"encoding/json"
	"github.com/google/uuid"
	"golang.org/x/net/websocket"
	"sync"
)

type User struct {
	UUID     uuid.UUID
	Nickname string
	parent   *Room
	conn     *websocket.Conn
	encoder  *json.Encoder
	cancel   func()
	once     sync.Once
}
