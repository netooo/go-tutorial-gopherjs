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

func (r *Room) do(ctx context.Context) {
	log.Println("start do room:", r.UUID)
	defer log.Println("stop do room:", r.UUID)
	members := map[string]*User{}
	defer func() {
		for _, m := range members {
			m.Close()
		}
		r.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case member := <-r.joinCh:
			join := &models.User{
				UUID:     member.UUID,
				Nickname: member.Nickname,
			}
			members[member.UUID.String()] = member
			for _, m := range members {
				if err := m.Write("join", join); err != nil {
					log.Println(err)
				}
			}
		case u := <-r.leaveCh:
			member := members[u.String()]
			if member != nil {
				leave := &models.User{
					UUID:     member.UUID,
					Nickname: member.Nickname,
				}
				for _, m := range members {
					if err := m.Write("leave", leave); err != nil {
						log.Println(err)
					}
				}
				member.Close()
			}
			delete(members, u.String())
		case msg := <-r.msgCh:
			for _, m := range members {
				if err := m.Write("message", msg); err != nil {
					log.Println(err)
				}
			}
		case ch := <-r.getUsersCh:
			res := make([]*models.User, 0, len(members))
			for _, m := range members {
				res = append(res, &models.User{
					UUID:     m.UUID,
					Nickname: m.Nickname,
				})
			}
			ch <- res
		}
		r.timer.Reset(TIMEOUT)
	}
}
