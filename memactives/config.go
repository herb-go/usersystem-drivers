package memactives

import (
	"errors"
	"time"

	"github.com/herb-go/usersystem"
)

type Config struct {
	Durations map[string]string
}

func (c *Config) CreateService() (*Service, error) {
	s := NewService()
	for ks := range c.Durations {
		k := usersystem.SessionType(ks)
		d, err := time.ParseDuration(c.Durations[ks])
		if err != nil {
			return nil, err
		}
		if d <= 0 {
			return nil, errors.New("duration must larger than 0")
		}
		s.Stores[k] = NewStoreList()
		store := NewStore()
		store.CreatedTime = time.Now()
		s.Stores[k].List = []*Store{store}
		s.Stores[k].Duration = d
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
