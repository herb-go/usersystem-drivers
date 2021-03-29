package userconfig

import (
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem-drivers/overseers/usersystemdirectivefactoryoverseer"
	"github.com/herb-go/worker"
)

type Directive struct {
	ID     string
	Config func(v interface{}) error `config:", lazyload"`
}

func (d *Directive) ApplyTo(s *usersystem.UserSystem) error {
	f := usersystemdirectivefactoryoverseer.GetUserSystemDirectiveFactoryByID(d.ID)
	if f == nil {
		return worker.NewWorkerNotFounderError(d.ID)
	}
	directive, err := f(d.Config)
	if err != nil {
		return err
	}
	return directive.Execute(s)
}

type Config struct {
	Directives []*Directive
}

func (c *Config) ApplyTo(s *usersystem.UserSystem) error {
	for k := range c.Directives {
		err := c.Directives[k].ApplyTo(s)
		if err != nil {
			return err
		}
	}
	return nil
}
