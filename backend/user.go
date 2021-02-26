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

func (u *User) Write(tp string, obj interface{}) error {
	return u.encoder.Encode(struct {
		Type string      `json:"type"`
		Data interface{} `json:"data"`
	}{tp, obj})
}

func (u *User) Close() {
	u.once.Do(func() {
		var conn *websocket.Conn
		conn, u.conn = u.conn, nil
		_ = conn.Close()
		u.cancel()
	})
}
