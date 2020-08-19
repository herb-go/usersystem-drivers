package herbsession

import (
	"net/http"

	"github.com/herb-go/session"
	"github.com/herb-go/usersystem"
)

type Service struct {
	Store session.Store
}

func (s *Service) GetSession(id string, st usersystem.SessionType) (usersystem.Session, error) {
	ts := s.Store.GetSession(id)
	us := NewSession()
	us.SessionType = st
	us.Session = ts
	return us, nil
}
func (s *Service) SessionMiddleware() func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return s.Store.InstallMiddleware()
}
func (s *Service) GetRequestSession(r *http.Request, st usersystem.SessionType) (usersystem.Session, error) {
	ts, err := s.Store.GetRequestSession(r)
	if err != nil {
		return nil, err
	}
	us := NewSession()
	us.SessionType = st
	us.Session = ts
	return us, nil
}

//Start start service
func (s *Service) Start() error {
	return nil
}

//Stop stop service
func (s *Service) Stop() error {
	return s.Store.Close()
}
