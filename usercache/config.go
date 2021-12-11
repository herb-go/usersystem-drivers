package usercache

import (
	"github.com/herb-go/datamodule-drivers/cacheconfig"
	"github.com/herb-go/datamodules/herbcache"
	"github.com/herb-go/datamodules/herbcache/cachepreset"
	"github.com/herb-go/herbdata/dataencoding/msgpackencoding"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/modules/useraccount"
	"github.com/herb-go/usersystem/modules/userstatus"
	"github.com/herb-go/usersystem/modules/userterm"
)

type Config struct {
	Cache         *cacheconfig.Config
	PrefixStatus  string
	PrefixTerm    string
	PrefixAccount string
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
	}
	if cache.PrefixStatus != "" {
		uss := userstatus.MustGetModule(s)
		if uss == nil {
			return ErrUserStatusServiceNotInstalled
		}
		us, err := cache.CreateStatus(uss.Service, cache.Preset)
		if err != nil {
			return err
		}
		uss.Service = us
	}
	if cache.PrefixTerm != "" {
		ust := userterm.MustGetModule(s)
		if ust == nil {
			return ErrUserTermServiceNotInstalled
		}
		us, err := cache.CreateTerm(ust.Service, cache.Preset)
		if err != nil {
			return err
		}
		ust.Service = us
	}
	if cache.PrefixAccount != "" {
		usa := useraccount.MustGetModule(s)
		if usa == nil {
			return ErrUserAccountServiceNotInstalled
		}
		us, err := cache.CreateAccount(usa.Service, cache.Preset)
		if err != nil {
			return err
		}
		usa.Service = us
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
