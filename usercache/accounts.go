package usercache

import (
	"github.com/herb-go/datamodules/herbcache/cachepreset"
	"github.com/herb-go/user"
	"github.com/herb-go/usersystem/modules/useraccount"
)

type Account struct {
	Cache *Cache
	useraccount.Service
	Preset *cachepreset.Preset
}

//Accounts return accounts of give uid.
func (a *Account) MustAccounts(uid string) *user.Accounts {
	var result = &user.Accounts{}
	err := a.Preset.Concat(cachepreset.Loader(func(id []byte) ([]byte, error) {
		result = a.Service.MustAccounts(string(id))
		return a.Preset.Encoding().Marshal(result)
	})).LoadS(uid, result)
	if err != nil {
		panic(err)
	}

	return result
}

//BindAccount bind account to user.
//If account exists,user.ErrAccountBindingExists should be rasied.
func (a *Account) MustBindAccount(uid string, account *user.Account) {
	a.Preset.DeleteS(uid)
	a.Service.MustBindAccount(uid, account)
}

//UnbindAccount unbind account from user.
//If account not exists,user.ErrAccountUnbindingNotExists should be rasied.
func (a *Account) MustUnbindAccount(uid string, account *user.Account) {
	a.Preset.DeleteS(uid)
	a.Service.MustUnbindAccount(uid, account)
}

//Start start service
func (a *Account) Start() error {
	a.Cache.Start()
	return a.Service.Start()
}

//Stop stop service
func (a *Account) Stop() error {
	a.Cache.Stop()
	return a.Service.Stop()
}

//Purge purge user data cache
func (a *Account) Purge(uid string) error {
	a.Preset.DeleteS(uid)
	return a.Service.Purge(uid)
}
