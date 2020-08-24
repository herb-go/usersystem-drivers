package memactives

import (
	"time"

	"github.com/herb-go/usersystem"
)

type Config struct {
	Durations map[usersystem.SessionType]time.Duration
}

func (c *Config) CreateService() (*Service, error) {
	s := NewService()
	for k := range c.Durations {
		s.Stores[k] = NewStoreList()
		store := NewStore()
		store.CreatedTime = time.Now()
		s.Stores[k].List = []*Store{store}
		s.Stores[k].Duration = c.Durations[k]
	}
	return s, nil
}

func (c *Config) Execute(us *usersystem.UserSystem) error {
	s, err := c.CreateService()
	if err != nil {
		return err
	}
	return s.Execute(us)
}

var DirectiveFactory = func(loader func(v interface{}) error) (usersystem.Directive, error) {
	c := &Config{}
	err := loader(c)

	if err != nil {
		return nil, err
	}

	return c, nil
}
