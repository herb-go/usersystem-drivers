package herbsession

import (
	"net/http"

	"github.com/herb-go/herbsecurity/authority"

	"github.com/herb-go/usersystem/httpusersystem/services/websession"

	"github.com/herb-go/herb/middleware"
	"github.com/herb-go/herbmodules/session"
	"github.com/herb-go/usersystem"
)

const PayloadsField = "payloads"

type Service struct {
	Prefix string
	Store  *session.Store
}

// Set set session by field name with given value.
func (s *Service) Set(r *http.Request, fieldname string, v interface{}) error {
	return s.Store.Set(r, s.Field(fieldname), v)
}

//Get get session by field name with given value.
func (s *Service) Get(r *http.Request, fieldname string, v interface{}) error {
	return s.Store.Get(r, s.Field(fieldname), v)
}

// Del del session value by field name .
func (s *Service) Del(r *http.Request, fieldname string) error {
	return s.Store.Del(r, s.Field(fieldname))
}

// IsNotFoundError check if given error is session not found error.
func (s *Service) IsNotFoundError(err error) bool {
	return s.Store.IsNotFoundError(err)
}
func (s *Service) PayloadsField() string {
	return s.Prefix + "." + PayloadsField
}
func (s *Service) Field(fieldname string) string {
	return s.Prefix + "-" + fieldname
}
func (s *Service) Execute(us *usersystem.UserSystem) error {
	hs, err := websession.GetService(us)
	if err != nil {
		return err
	}
	if hs == nil {
		return nil
	}
	if s.Prefix == "" {
		s.Prefix = string(us.Keyword)
	}

	hs.Service = s
	return nil
}
func (s *Service) GetSession(st usersystem.SessionType, id string) (*usersystem.Session, error) {
	ts := s.Store.GetSession(id)
	payloads := authority.NewPayloads()
	err := ts.Get(s.PayloadsField(), &payloads)
	if err != nil {
		if s.Store.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return usersystem.NewSession().WithType(st).WithPayloads(payloads).WithID(id), nil
}

func (s *Service) RevokeSession(code string) (bool, error) {
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
	token, err := ts.Token()
	if err != nil {
		return nil, err
	}
	payloads := authority.NewPayloads()
	err = ts.Get(s.PayloadsField(), &payloads)
	if err != nil {
		if s.Store.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}
	return usersystem.NewSession().WithType(st).WithPayloads(payloads).WithID(token), nil

}

func (s *Service) LoginRequestSession(r *http.Request, payloads *authority.Payloads) (*usersystem.Session, error) {
	var id string
	ts, err := s.Store.GetRequestSession(r)
	if err != nil {
		return nil, err
	}
	if s.Store.Driver.DynamicToken() {
		err = s.Store.Set(r, s.PayloadsField(), payloads)
		if err != nil {
			return nil, err
		}
		id, err = ts.Token()
		if err != nil {
			return nil, err
		}
	} else {
		id, err = ts.Token()
		if err != nil {
			return nil, err
		}
		payloads.Set(usersystem.PayloadRevokeCode, []byte(id))
		err = s.Store.Set(r, s.PayloadsField(), payloads)
		if err != nil {
			return nil, err
		}
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
