package redisactives

import (
	"errors"
	"time"

	"github.com/herb-go/datasource/redis/redispool"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/modules/activesessions"
)

type Config struct {
	redispool.Config
	Interval  int64
	Prefix    string
	Durations map[string]string
}

func (c *Config) CreateService() (*Service, error) {
	s := NewService()
	p := redispool.New()
	err := c.Config.ApplyTo(p)
	if err != nil {
		return nil, err
	}
	s.Pool = p.Open()
	for ks := range c.Durations {
		k := usersystem.SessionType(ks)
		d, err := time.ParseDuration(c.Durations[ks])
		if err != nil {
			return nil, err
		}
		if d <= 0 {
			return nil, errors.New("duration must larger than 0")
		}
		s.configs[k] = &activesessions.Config{
			Duration:  d,
			Supported: true,
		}
	}
	s.Interval = c.Interval
	s.Prefix = c.Prefix
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
