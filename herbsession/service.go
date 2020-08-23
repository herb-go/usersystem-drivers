package herbsession

import (
	"net/http"

	"github.com/herb-go/herbsecurity/authority"

	"github.com/herb-go/usersystem/httpusersystem/services/httpsession"

	"github.com/herb-go/herb/middleware"
	"github.com/herb-go/session"
	"github.com/herb-go/usersystem"
)

type Service struct {
	Prefix string
	Store  *session.Store
}

func (s *Service) PayloadsField() string {
	return s.Prefix + "." + PayloadsField
}

func (s *Service) Execute(us *usersystem.UserSystem) error {
	v, err := us.GetConfigurableService(httpsession.ServiceName)
	if err != nil {
		return err
	}
	if v == nil {
		return nil
	}
	hs := v.(*httpsession.HTTPSession)
	hs.Service = s
	return nil
}
func (s *Service) GetSession(st usersystem.SessionType, id string) (*usersystem.Session, error) {
	ts := s.Store.GetSession(id)
	payloads := authority.NewPayloads()
	err := ts.Get(s.PayloadsField(), &payloads)
	if err != nil {
		return nil, err
	}
	return usersystem.NewSession().WithType(st).WithPayloads(payloads), nil
}

func (s *Service) RevokeSession(st usersystem.SessionType, code string) (bool, error) {
	return s.Store.Driver.Delete(code)
}
func (s *Service) SessionMiddleware() func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return middleware.New(s.Store.InstallMiddleware(), s.Store.AutoGenerateMiddleware()).ServeMiddleware
}
func (s *Service) GetRequestSession(r *http.Request, st usersystem.SessionType) (*usersystem.Session, error) {
	ts, err := s.Store.GetRequestSession(r)
	if err != nil {
		return nil, err
	}
	payloads := authority.NewPayloads()
	err = ts.Get(s.PayloadsField(), &payloads)
	if err != nil {
		return nil, err
	}
	return usersystem.NewSession().WithType(st).WithPayloads(payloads), nil

}

func (s *Service) LoginRequestSession(r *http.Request, payloads *authority.Payloads) (*usersystem.Session, error) {
	err := s.Store.Set(r, s.PayloadsField(), payloads)
	if err != nil {
		return nil, err
	}
	ts, err := s.Store.GetRequestSession(r)
	if err != nil {
		return nil, err
	}
	id, err := ts.Token()
	if err != nil {
		return nil, err
	}
	return usersystem.NewSession().WithID(id).WithPayloads(payloads.Clone()), nil
}
func (s *Service) LogoutRequestSession(r *http.Request) (bool, error) {
	return true, s.Store.Del(r, s.PayloadsField())
}

//Start start service
func (s *Service) Start() error {
	return nil
}

//Stop stop service
func (s *Service) Stop() error {
	return s.Store.Close()
}

func NewService() *Service {
	return &Service{}
}
