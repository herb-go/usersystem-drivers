package herbsession

import (
	"github.com/herb-go/deprecated/session"
	"github.com/herb-go/usersystem"
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

func (c *Config) Execute(s *usersystem.UserSystem) error {
	service, err := c.CreateService()
	if err != nil {
		return err
	}
	return service.Execute(s)
}

var DirectiveFactory = func(loader func(v interface{}) error) (usersystem.Directive, error) {
	c := &Config{}
	err := loader(c)

	if err != nil {
		return nil, err
	}

	return c, nil
}
