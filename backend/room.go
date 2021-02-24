package backend

import (
	"context"
	"github.com/google/uuid"
	"github.com/netooo/go-tutorial-gopherjs/app/models"
	"log"
	"sync"
	"time"
)

type Room struct {
	UUID       uuid.UUID
	joinCh     chan *User
	leaveCh    chan uuid.UUID
	msgCh      chan *models.Message
	getUsersCh chan chan []*models.User
	timer      *time.Timer
	cancel     func()
	once       sync.Once
}

var (
	newRoomCh = make(chan *models.Room)
	delRoomCh = make(chan uuid.UUID, 1)
	getRoomCh = make(chan getRoom)
)

type getRoom struct {
	uuid   string
	result chan *Room
}

func roomManage(ctx context.Context) {
	rooms := map[string]*Room{}
	for {
		select {
		case <-ctx.Done():
			return
		case u := <-delRoomCh:
			rooms[u.String()].Close()
			delete(rooms, u.String())
			log.Println("del room:", u.String())
		case room := <-newRoomCh:
			rooms[room.UUID.String()] = NewRoom(context.Background(), room)
			log.Println("new room:", room.UUID.String())
		case req := <-getRoomCh:
			req.result <- rooms[req.uuid]
		}
	}
}
