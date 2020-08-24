package tomluser

import (
	"sync"

	"github.com/herb-go/usersystem/services/userprofile"

	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/services/useraccount"
	"github.com/herb-go/usersystem/services/userpassword"
	"github.com/herb-go/usersystem/services/userrole"
	"github.com/herb-go/usersystem/services/userstatus"
	"github.com/herb-go/usersystem/services/userterm"

	"github.com/herb-go/providers/herb/statictoml"
)

var locker sync.Mutex
var registered = map[statictoml.Source]*Users{}

func Flush() {
	locker.Lock()
	locker.Unlock()
	registered = map[statictoml.Source]*Users{}
}

type Data struct {
	Users []*User
}

func NewData() *Data {
	return &Data{}
}

type Config struct {
	Source        statictoml.Source
	ProfileFields []string
	ServePassword bool
	ServeStatus   bool
	ServeAccounts bool
	ServeRoles    bool
	ServeTerm     bool
	ServeProfile  bool
	HashMode      string
}

func (c *Config) Load() (*Users, error) {
	locker.Lock()
	locker.Unlock()
	source, err := c.Source.Abs()
	if err != nil {
		return nil, err
	}
	u, ok := registered[source]
	if ok && u != nil {
		return u, nil
	}
	u = NewUsers()
	u.Source = c.Source
	data := NewData()
	err = u.Source.Load(data)
	if err != nil {
		return nil, err
	}
	for _, v := range c.ProfileFields {
		u.ProfileFields[v] = true
	}
	for k := range data.Users {
		u.addUser(data.Users[k])
	}
	return u, nil
}
func (c *Config) Execute(s *usersystem.UserSystem) error {
	u, err := c.Load()
	if err != nil {
		return err
	}
	if c.ServeStatus {
		ss, err := userstatus.GetService(s)
		if err != nil {
			return err
		}
		if ss != nil {
			ss.Service = u
		}
	}
	if c.ServeAccounts {
		ua, err := useraccount.GetService(s)
		if err != nil {
			return err
		}
		if ua != nil {
			ua.Service = u
		}
	}
	if c.ServePassword {
		up, err := userpassword.GetService(s)
		if err != nil {
			return err
		}
		if up != nil {
			up.Service = u
		}
	}
	if c.ServeRoles {
		ur, err := userrole.GetService(s)
		if err != nil {
			return err
		}
		if ur != nil {
			ur.Service = u
		}
	}
	if c.ServeTerm {
		ut, err := userterm.GetService(s)
		if err != nil {
			return err
		}
		if ut != nil {
			ut.Service = u
		}
	}
	if c.ServeProfile {
		up, err := userprofile.GetService(s)
		if err != nil {
			return err
		}
		if up != nil {
			up.AppendService(u)
		}
	}
	return nil
}

var DirectiveFactory = func(loader func(v interface{}) error) (usersystem.Directive, error) {
	c := &Config{}
	err := loader(c)

	if err != nil {
		return nil, err
	}

	return c, nil
}
