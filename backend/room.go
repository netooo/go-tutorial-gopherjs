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

func GetRoom(uid string) *Room {
	ch := make(chan *Room, 1)
	getRoomCh <- getRoom{uid, ch}
	return <-ch
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

func (r *Room) Close() {
	r.once.Do(func() {
		r.cancel()
	})
}

func (r *Room) do(ctx context.Context) {
	log.Println("start do room:", r.UUID)
	defer log.Println("stop do room:", r.UUID)
	users := map[string]*User{}
	defer func() {
		for _, u := range users {
			u.Close()
		}
		r.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			return
		case user := <-r.joinCh:
			join := &models.User{
				UUID:     user.UUID,
				Nickname: user.Nickname,
			}
			users[user.UUID.String()] = user
			for _, u := range users {
				if err := u.Write("join", join); err != nil {
					log.Println(err)
				}
			}
		case u := <-r.leaveCh:
			user := users[u.String()]
			if user != nil {
				leave := &models.User{
					UUID:     user.UUID,
					Nickname: user.Nickname,
				}
				for _, u := range users {
					if err := u.Write("leave", leave); err != nil {
						log.Println(err)
					}
				}
				user.Close()
			}
			delete(users, u.String())
		case msg := <-r.msgCh:
			for _, u := range users {
				if err := u.Write("message", msg); err != nil {
					log.Println(err)
				}
			}
		case ch := <-r.getUsersCh:
			res := make([]*models.User, 0, len(users))
			for _, u := range users {
				res = append(res, &models.User{
					UUID:     u.UUID,
					Nickname: u.Nickname,
				})
			}
			ch <- res
		}
		r.timer.Reset(TIMEOUT)
	}
}
