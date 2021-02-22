package backend

import (
	"github.com/google/uuid"
	"github.com/netooo/go-tutorial-gopherjs/app/models"
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
