package usercache

import (
	"github.com/herb-go/datamodules/herbcache/cachepreset"
	"github.com/herb-go/user/status"
	"github.com/herb-go/usersystem/modules/userstatus"
)

type cachedStatus struct {
	Status int
	Exists bool
}

type Status struct {
	Cache *Cache
	userstatus.Service
	Preset *cachepreset.Preset
}

func (s *Status) MustLoadStatus(id string) (status.Status, bool) {
	var result = &cachedStatus{}
	err := s.Preset.Concat(cachepreset.Loader(func(id []byte) ([]byte, error) {
		st, exists := s.Service.MustLoadStatus(string(id))
		result.Status = int(st)
		result.Exists = exists
		return s.Preset.Encoding().Marshal(result)
	})).LoadS(id, result)
	if err != nil {
		panic(err)
	}
	return status.Status(result.Status), result.Exists
}

//MustUpdateStatus update user status.
func (s *Status) MustUpdateStatus(id string, st status.Status) {
	s.Preset.DeleteS(id)
	s.Service.MustUpdateStatus(id, st)

}

//MustRemoveStatus remove user status
func (s *Status) MustRemoveStatus(id string) {
	s.Preset.DeleteS(id)
	s.Service.MustRemoveStatus(id)
}

//MustRemoveStatus list user by status
//Purge purge user data cache
func (s *Status) Purge(id string) error {
	s.Preset.DeleteS(id)

	return s.Service.Purge(id)
}

//Start start service
func (s *Status) Start() error {
	s.Cache.Start()
	return s.Service.Start()
}

//Stop stop service
func (s *Status) Stop() error {
	s.Cache.Stop()
	return s.Service.Stop()
}
