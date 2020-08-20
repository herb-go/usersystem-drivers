package herbsession

import (
	"github.com/herb-go/session"
)

type Config struct {
	Prefix string
	*session.StoreConfig
}

func (c *Config) CreateService() (*Service, error) {
	s := NewService()
	s.Prefix = c.Prefix
	store := session.New()
	err := c.StoreConfig.ApplyTo(store)
	if err != nil {
		return nil, err
	}
	s.Store = store
	return s, nil
}
