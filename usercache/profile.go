package usercache

import (
	"github.com/herb-go/datamodules/herbcache/cachepreset"
	"github.com/herb-go/user/profile"
	"github.com/herb-go/usersystem/modules/userprofile"
)

type Profile struct {
	Cache *Cache
	userprofile.Service
	Preset *cachepreset.Preset
}

func (p *Profile) MustGetProfile(id string) *profile.Profile {
	var result = &profile.Profile{}
	err := p.Preset.Concat(cachepreset.Loader(func(id []byte) ([]byte, error) {
		result = p.Service.MustGetProfile(string(id))
		return p.Preset.Encoding().Marshal(result)
	})).LoadS(id, &result)
	if err != nil {
		panic(err)
	}
	return result
}
func (p *Profile) MustUpdateProfile(id string, pf *profile.Profile) {
	p.Preset.DeleteS(id)
	p.Service.MustUpdateProfile(id, pf)
}

//Start start service
func (p *Profile) Start() error {
	p.Cache.Start()
	return p.Service.Start()

}

//Stop stop service
func (p *Profile) Stop() error {
	p.Cache.Stop()
	return p.Service.Stop()

}

//Purge purge user data cache
func (p *Profile) Purge(uid string) error {
	p.Preset.DeleteS(uid)
	return p.Service.Purge(uid)
}
