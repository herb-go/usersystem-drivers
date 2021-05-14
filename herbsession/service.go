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
	hs := websession.MustGetModule(us)

	if hs == nil {
		return nil
	}
	if len(s.Prefix) == 0 {
		s.Prefix = []byte(us.Keyword)
	}

	hs.Service = s
	return nil
}
func (s *Service) MustGetSession(st usersystem.SessionType, id string) *usersystem.Session {
	session, err := s.Store.LoadSession(id)
	if err != nil {
		if err == herbdata.ErrNotFound {
			return nil
		}
		panic(err)
	}
	data, err := session.Get(s.payloadsField())
	if err != nil {
		if err == herbdata.ErrNotFound {
			return nil
		}
		panic(err)
	}
	payloads := authority.NewPayloads()
	err = msgpack.Unmarshal(data, &payloads)
	if err != nil {
		panic(err)
	}
	return usersystem.NewSession().WithType(st).WithPayloads(payloads).WithID(id)
}

func (s *Service) MustRevokeSession(code string) bool {
	ok := !s.Store.Engine.DynamicToken()
	err := s.Store.RevokeSession(code)
	if err != nil {
		panic(err)
	}
	return ok
}
func (s *Service) SessionMiddleware() func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	return s.Store.Install()
}
func (s *Service) MustGetRequestSession(r *http.Request, st usersystem.SessionType) *usersystem.Session {
	session := s.Store.RequestSession(r)
	token := session.Token()
	payloads := authority.NewPayloads()
	data, err := session.Get(s.payloadsField())
	if err != nil {
		if err == herbdata.ErrNotFound {
			return nil
		}
		panic(err)
	}
	err = msgpack.Unmarshal(data, &payloads)
	if err != nil {
		panic(err)
	}
	return usersystem.NewSession().WithType(st).WithPayloads(payloads).WithID(token)

}

func (s *Service) MustLoginRequestSession(r *http.Request, payloads *authority.Payloads) *usersystem.Session {
	var err error
	var data []byte
	session := s.Store.RequestSession(r)
	if !s.Store.Engine.DynamicToken() {
		payloads.Set(usersystem.PayloadRevokeCode, []byte(session.Token()))
	}
	data, err = msgpack.Marshal(payloads)

	err = session.Set(s.payloadsField(), data)
	if err != nil {
		panic(err)
	}
	if s.Store.Engine.DynamicToken() {
		err = s.Store.SaveSession(session)
		if err != nil {
			panic(err)
		}
	}
	return usersystem.NewSession().WithID(session.Token()).WithPayloads(payloads.Clone())
}
func (s *Service) MustLogoutRequestSession(r *http.Request) bool {
	session := s.Store.RequestSession(r)
	err := session.Delete(s.payloadsField())
	if err != nil {
		panic(err)
	}
	return true
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
