package backend

import (
	"context"
	"github.com/google/uuid"
	"github.com/netooo/go-tutorial-gopherjs/app/models"
	"log"
	"sync"
	"time"
)

const TIMEOUT = 300 * time.Second

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

func NewRoom(ctx context.Context, room *models.Room) *Room {
	r := &Room{
		UUID:       room.UUID,
		joinCh:     make(chan *User),
		leaveCh:    make(chan uuid.UUID),
		msgCh:      make(chan *models.Message),
		getUsersCh: make(chan chan []*models.User),
		timer: time.AfterFunc(TIMEOUT, func() {
			delRoomCh <- room.UUID
		}),
	}
	ctx, r.cancel = context.WithCancel(ctx)
	go r.do(ctx)
	return r
}
