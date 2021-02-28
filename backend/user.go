package backend

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/netooo/go-tutorial-gopherjs/app/models"
	"golang.org/x/net/websocket"
	"io"
	"log"
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

func NewUser(ctx context.Context, parent *Room, user *models.User, conn *websocket.Conn) *User {
	u := &User{
		UUID:     user.UUID,
		Nickname: user.Nickname,
		parent:   parent,
		conn:     conn,
		encoder:  json.NewEncoder(conn),
	}
	ctx, u.cancel = context.WithCancel(ctx)
	go u.do(ctx)
	return u
}

func (u *User) do(ctx context.Context) {
	defer u.Close()
	defer u.parent.Leave(u.UUID)
	reader := json.NewDecoder(u.conn)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		var v *models.Event
		if err := reader.Decode(&v); err != nil {
			if err == io.EOF {
				return
			}
			log.Println(err)
			continue
		}
		log.Println("received:", v.Type, string(v.Data))
		switch v.Type {
		case "message":
			var message *models.Message
			if err := v.Unmarshal(&message); err != nil {
				log.Println(err)
				continue
			}
			u.parent.Publish(message)
		}
	}
}
