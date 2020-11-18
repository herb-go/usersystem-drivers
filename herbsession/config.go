package herbsession

import (
	"github.com/herb-go/herbmodules/httpsession"
	"github.com/herb-go/usersystem"
)

type Config struct {
	Prefix string
	*httpsession.Config
}

func (c *Config) CreateService() (*Service, error) {
	s := NewService()
	s.Prefix = []byte(c.Prefix)
	store := httpsession.New()
	err := c.Config.ApplyTo(store)
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
