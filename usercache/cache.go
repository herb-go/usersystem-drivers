package usercache

import (
	"sync"

	"github.com/herb-go/datamodules/herbcache"
	"github.com/herb-go/datamodules/herbcache/cachepreset"
	"github.com/herb-go/usersystem/modules/useraccount"
	"github.com/herb-go/usersystem/modules/userstatus"
	"github.com/herb-go/usersystem/modules/userterm"
)

type Cache struct {
	Stroage       *herbcache.Storage
	Preset        *cachepreset.Preset
	PrefixStatus  string
	PrefixAccount string
	PrefixTerm    string
	started       bool
	lock          sync.Mutex
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
