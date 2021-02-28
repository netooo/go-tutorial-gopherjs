package backend

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/netooo/go-tutorial-gopherjs/app/models"
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

func NewUser(ctx context.Context, parent *Room, member *models.User, conn *websocket.Conn) *User {
	m := &User{
		UUID:     member.UUID,
		Nickname: member.Nickname,
		parent:   parent,
		conn:     conn,
		encoder:  json.NewEncoder(conn),
	}
	ctx, m.cancel = context.WithCancel(ctx)
	go m.do(ctx)
	return m
}
