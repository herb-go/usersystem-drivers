package usercache

import (
	"github.com/herb-go/datamodules/herbcache/cachepreset"
	"github.com/herb-go/herbsecurity/authorize/role"
	"github.com/herb-go/usersystem/modules/userrole"
)

type Role struct {
	Cache *Cache
	userrole.Service
	Preset *cachepreset.Preset
}

func (r *Role) MustRoles(uid string) *role.Roles {
	var result = &role.Roles{}
	err := r.Preset.Concat(cachepreset.Loader(func(id []byte) ([]byte, error) {
		result = r.Service.MustRoles(string(id))
		return r.Preset.Encoding().Marshal(result)
	})).LoadS(uid, result)
	if err != nil {
		panic(err)
	}
	return result
}

//Start start service
func (r *Role) Start() error {
	r.Cache.Start()
	return r.Service.Start()

}

//Stop stop service
func (r *Role) Stop() error {
	r.Cache.Stop()
	return r.Service.Stop()
}

//Purge purge user data cache
func (r *Role) Purge(uid string) error {
	r.Preset.DeleteS(uid)
	return r.Service.Purge(uid)
}
