package usercache

import (
	"github.com/herb-go/datamodule-drivers/cacheconfig"
	"github.com/herb-go/datamodules/herbcache"
	"github.com/herb-go/datamodules/herbcache/cachepreset"
	"github.com/herb-go/herbdata/dataencoding/msgpackencoding"
	"github.com/herb-go/usersystem"
)

type Config struct {
	Cache         *cacheconfig.Config
	PrefixStatus  string
	PrefixTerm    string
	PrefixAccount string
	PrefixRole    string
	PrefixProfile string
}

func (c *Config) Execute(s *usersystem.UserSystem) error {
	storage := herbcache.NewStorage()
	err := c.Cache.Storage.ApplyTo(storage)
	if err != nil {
		return err
	}
	hc := herbcache.New().OverrideStorage(storage)
	preset, err := c.Cache.Preset.Exec(cachepreset.New(cachepreset.Cache(hc), cachepreset.Encoding(msgpackencoding.Encoding)))
	cache := &Cache{
		Stroage:       storage,
		Preset:        preset,
		PrefixStatus:  c.PrefixStatus,
		PrefixTerm:    c.PrefixTerm,
		PrefixAccount: c.PrefixAccount,
		PrefixRole:    c.PrefixRole,
		PrefixProfile: c.PrefixProfile,
	}

	return cache.Execute(s)
}

var DirectiveFactory = func(loader func(v interface{}) error) (usersystem.Directive, error) {
	c := &Config{}
	err := loader(c)

	if err != nil {
		return nil, err
	}

	return c, nil
}
