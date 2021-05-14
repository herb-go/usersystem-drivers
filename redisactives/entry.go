package redisactives

import (
	"time"

	"github.com/herb-go/usersystem/modules/activesessions"
)

type entry struct {
	ID         string
	LastActive int64
}

func (e *entry) Convert() *activesessions.Active {
	return &activesessions.Active{
		SessionID:  e.ID,
		LastActive: time.Unix(e.LastActive, 0),
	}
}
