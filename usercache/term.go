package usercache

import (
	"github.com/herb-go/datamodules/herbcache/cachepreset"
	"github.com/herb-go/usersystem/modules/userterm"
)

type Term struct {
	Cache *Cache
	userterm.Service
	Preset *cachepreset.Preset
}

func (t *Term) MustCurrentTerm(uid string) string {
	var result = ""
	err := t.Preset.Concat(cachepreset.Loader(func(id []byte) ([]byte, error) {
		result = t.Service.MustCurrentTerm(string(id))
		return t.Preset.Encoding().Marshal(result)
	})).LoadS(uid, &result)
	if err != nil {
		panic(err)
	}

	return result
}
func (t *Term) MustStartNewTerm(uid string) string {
	t.Preset.DeleteS(uid)
	return t.Service.MustStartNewTerm(uid)
}

//Start start service
func (t *Term) Start() error {
	t.Cache.Start()
	return t.Service.Start()
}

//Stop stop service
func (t *Term) Stop() error {
	t.Cache.Stop()
	return t.Service.Stop()
}

//Purge purge user data cache
func (t *Term) Purge(uid string) error {
	t.Preset.DeleteS(uid)
	return t.Service.Purge(uid)
}
