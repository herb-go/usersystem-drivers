package usercache

import (
	"sync"

	"github.com/herb-go/datamodules/herbcache"
	"github.com/herb-go/datamodules/herbcache/cachepreset"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/modules/useraccount"
	"github.com/herb-go/usersystem/modules/userprofile"
	"github.com/herb-go/usersystem/modules/userrole"
	"github.com/herb-go/usersystem/modules/userstatus"
	"github.com/herb-go/usersystem/modules/userterm"
)

type Cache struct {
	Stroage       *herbcache.Storage
	Preset        *cachepreset.Preset
	PrefixStatus  string
	PrefixAccount string
	PrefixTerm    string
	PrefixRole    string
	PrefixProfile string
	started       bool
	lock          sync.Mutex
}

func (c *Cache) Execute(s *usersystem.UserSystem) error {
	if c.PrefixStatus != "" {
		uss := userstatus.MustGetModule(s)
		if uss == nil {
			return ErrUserStatusServiceNotInstalled
		}
		us, err := c.CreateStatus(uss.Service, c.Preset)
		if err != nil {
			return err
		}
		uss.Service = us
	}
	if c.PrefixTerm != "" {
		ust := userterm.MustGetModule(s)
		if ust == nil {
			return ErrUserTermServiceNotInstalled
		}
		us, err := c.CreateTerm(ust.Service, c.Preset)
		if err != nil {
			return err
		}
		ust.Service = us
	}
	if c.PrefixAccount != "" {
		usa := useraccount.MustGetModule(s)
		if usa == nil {
			return ErrUserAccountServiceNotInstalled
		}
		us, err := c.CreateAccount(usa.Service, c.Preset)
		if err != nil {
			return err
		}
		usa.Service = us
	}
	if c.PrefixRole != "" {
		usr := userrole.MustGetModule(s)
		if usr == nil {
			return ErrUserRoleServiceNotInstalled
		}
		us, err := c.CreateRole(usr.Service, c.Preset)
		if err != nil {
			return err
		}
		usr.Service = us
	}
	if c.PrefixProfile != "" {
		usp := userprofile.MustGetModule(s)
		if usp == nil {
			return ErrUserRoleServiceNotInstalled
		}
		us, err := c.CreateProfile(usp.Services, c.Preset)
		if err != nil {
			return err
		}
		usp.Services = userprofile.Services{us}
	}
	return nil

}
func (c *Cache) Start() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.started {
		return nil
	}
	c.started = true
	return c.Stroage.Start()
}
func (c *Cache) Stop() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if !c.started {
		return nil
	}
	c.started = false
	return c.Stroage.Stop()
}
func (c *Cache) CreateStatus(service userstatus.Service, preset *cachepreset.Preset) (*Status, error) {
	p, err := c.Preset.Concat(cachepreset.ChildCache(c.PrefixStatus), cachepreset.Flushable(true)).Apply()
	if err != nil {
		return nil, err
	}
	return &Status{
		Cache:   c,
		Service: service,
		Preset:  p,
	}, nil
}

func (c *Cache) CreateTerm(service userterm.Service, preset *cachepreset.Preset) (*Term, error) {
	p, err := c.Preset.Concat(cachepreset.ChildCache(c.PrefixTerm), cachepreset.Flushable(true)).Apply()
	if err != nil {
		return nil, err
	}
	return &Term{
		Cache:   c,
		Service: service,
		Preset:  p,
	}, nil
}

func (c *Cache) CreateAccount(service useraccount.Service, preset *cachepreset.Preset) (*Account, error) {
	p, err := c.Preset.Concat(cachepreset.ChildCache(c.PrefixAccount), cachepreset.Flushable(true)).Apply()
	if err != nil {
		return nil, err
	}
	return &Account{
		Cache:   c,
		Service: service,
		Preset:  p,
	}, nil
}

func (c *Cache) CreateRole(service userrole.Service, preset *cachepreset.Preset) (*Role, error) {
	p, err := c.Preset.Concat(cachepreset.ChildCache(c.PrefixRole), cachepreset.Flushable(true)).Apply()
	if err != nil {
		return nil, err
	}
	return &Role{
		Cache:   c,
		Service: service,
		Preset:  p,
	}, nil
}

func (c *Cache) CreateProfile(service userprofile.Service, preset *cachepreset.Preset) (*Profile, error) {
	p, err := c.Preset.Concat(cachepreset.ChildCache(c.PrefixProfile), cachepreset.Flushable(true)).Apply()
	if err != nil {
		return nil, err
	}
	return &Profile{
		Cache:   c,
		Service: service,
		Preset:  p,
	}, nil
}
