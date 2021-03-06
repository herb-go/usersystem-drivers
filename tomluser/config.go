package tomluser

import (
	"sync"

	"github.com/herb-go/usersystem/modules/userprofile"

	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/modules/useraccount"
	"github.com/herb-go/usersystem/modules/userpassword"
	"github.com/herb-go/usersystem/modules/userrole"
	"github.com/herb-go/usersystem/modules/userstatus"
	"github.com/herb-go/usersystem/modules/userterm"

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
	Example       statictoml.Source
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
	err = source.VerifyWithExample(c.Example)
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
		ss := userstatus.MustGetModule(s)
		if ss != nil {
			ss.Service = u
		}
	}
	if c.ServeAccounts {
		ua := useraccount.MustGetModule(s)
		if ua != nil {
			ua.Service = u
		}
	}
	if c.ServePassword {
		up := userpassword.MustGetModule(s)
		if up != nil {
			up.Service = u
		}
	}
	if c.ServeRoles {
		ur := userrole.MustGetModule(s)
		if ur != nil {
			ur.Service = u
		}
	}
	if c.ServeTerm {
		ut := userterm.MustGetModule(s)
		if ut != nil {
			ut.Service = u
		}
	}
	if c.ServeProfile {
		up := userprofile.MustGetModule(s)
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
