package herbsession

import (
	"net/http"

	"github.com/herb-go/herbdata/datautil"

	"github.com/herb-go/herbdata"

	"github.com/herb-go/herbmodules/httpsession"
	"github.com/herb-go/herbsecurity/authority"
	"github.com/vmihailenco/msgpack"

	"github.com/herb-go/usersystem/httpusersystem/services/websession"

	"github.com/herb-go/usersystem"
)

var PayloadsField = []byte("payloads")

type Service struct {
	Prefix []byte
	Store  *httpsession.Store
}

// Set set session by field name with given value.
func (s *Service) Set(r *http.Request, fieldname string, v interface{}) error {
	session := s.Store.RequestSession(r)
	data, err := msgpack.Marshal(v)
	if err != nil {
		return err
	}
	return session.Set(s.field(fieldname), data)
}

//Get get session by field name with given value.
func (s *Service) Get(r *http.Request, fieldname string, v interface{}) error {
	session := s.Store.RequestSession(r)
	data, err := session.Get(s.field(fieldname))
	if err != nil {
		return err
	}
	return msgpack.Unmarshal(data, v)
}

// Del del session value by field name .
func (s *Service) Del(r *http.Request, fieldname string) error {
	session := s.Store.RequestSession(r)
	return session.Delete(s.field(fieldname))
}

// IsNotFoundError check if given error is session not found error.
func (s *Service) IsNotFoundError(err error) bool {
	return err == herbdata.ErrNotFound
}
func (s *Service) payloadsField() []byte {
	return datautil.Join(nil, s.Prefix, []byte("."), PayloadsField)

}
func (s *Service) field(fieldname string) []byte {
	return datautil.Join(nil, []byte(s.Prefix), []byte{}, []byte(fieldname))
}
func (s *Service) Execute(us *usersystem.UserSystem) error {
	hs, err := websession.GetService(us)
	if err != nil {
		return err
	}
	if hs == nil {
		return nil
	}
	if len(s.Prefix) == 0 {
		s.Prefix = []byte(us.Keyword)
	}

	hs.Service = s
	return nil
}
func (s *Service) GetSession(st usersystem.SessionType, id string) (*usersystem.Session, error) {
	session, err := s.Store.LoadSession(id)
	if err != nil {
		if err == herbdata.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	data, err := session.Get(s.payloadsField())
	if err != nil {
		if err == herbdata.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	payloads := authority.NewPayloads()
	err = msgpack.Unmarshal(data, &payloads)
	if err != nil {
		return nil, err
	}
	return usersystem.NewSession().WithType(st).WithPayloads(payloads).WithID(id), nil
}

func (s *Service) RevokeSession(code string) (bool, error) {
	return !s.Store.Engine.DynamicToken(), s.Store.RevokeSession(code)
}
func (s *Service) SessionMiddleware() func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return s.Store.Install()
}
func (s *Service) GetRequestSession(r *http.Request, st usersystem.SessionType) (*usersystem.Session, error) {
	session := s.Store.RequestSession(r)
	token := session.Token()
	payloads := authority.NewPayloads()
	data, err := session.Get(s.payloadsField())
	if err != nil {
		if err == herbdata.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	err = msgpack.Unmarshal(data, &payloads)
	if err != nil {
		return nil, err
	}
	return usersystem.NewSession().WithType(st).WithPayloads(payloads).WithID(token), nil

}

func (s *Service) LoginRequestSession(r *http.Request, payloads *authority.Payloads) (*usersystem.Session, error) {
	var err error
	var data []byte
	session := s.Store.RequestSession(r)
	data, err = msgpack.Marshal(payloads)
	if s.Store.Engine.DynamicToken() {
		if err != nil {
			return nil, err
		}
		err = s.Store.SaveSession(session)
		if err != nil {
			return nil, err
		}

	} else {
		payloads.Set(usersystem.PayloadRevokeCode, []byte(session.Token()))

		err = session.Set(s.payloadsField(), data)
		if err != nil {
			return nil, err
		}
	}
	return usersystem.NewSession().WithID(session.Token()).WithPayloads(payloads.Clone()), nil
}
func (s *Service) LogoutRequestSession(r *http.Request) (bool, error) {
	session := s.Store.RequestSession(r)
	return true, session.Delete(s.payloadsField())
}

//Start start service
func (s *Service) Start() error {
	return s.Store.Start()
}

//Stop stop service
func (s *Service) Stop() error {
	return s.Store.Stop()
}

func NewService() *Service {
	return &Service{}
}
