package herbsession

import (
	"github.com/herb-go/herb/cache"
	"github.com/herb-go/herbsecurity/authority"
	"github.com/herb-go/session"
	"github.com/herb-go/usersystem"
)

const UIDField = "uid"
const PayloadsField = "payloads"
const ValuePrefix = "."

type Session struct {
	Prefix      string
	Session     *session.Session
	SessionType usersystem.SessionType
}

func (s *Session) fieldname(fieldname string, passthrough bool) string {
	if passthrough {
		return s.Prefix + cache.KeyPrefix + fieldname
	}
	return s.Prefix + cache.KeyPrefix + ValuePrefix + fieldname
}
func (s *Session) ID() string {
	return s.Session.MustToken()
}
func (s *Session) Type() usersystem.SessionType {
	return s.SessionType
}
func (s *Session) UID() (string, error) {
	uid := ""
	err := s.Session.Get(s.fieldname(UIDField, true), &uid)
	if err != nil {
		return "", err
	}
	return uid, nil
}
func (s *Session) SaveUID(uid string) error {
	return s.Session.Set(s.fieldname(UIDField, true), uid)
}
func (s *Session) Payloads() (*authority.Payloads, error) {
	p := authority.NewPayloads()
	err := s.Session.Get(s.fieldname(PayloadsField, true), p)
	if err != nil {
		return nil, err
	}
	return p, nil

}
func (s *Session) SavePayloads(p *authority.Payloads) error {
	return s.Session.Set(s.fieldname(PayloadsField, true), p)
}
func (s *Session) Destory() (bool, error) {
	return s.Session.Destory()
}
func (s *Session) Save(key string, v interface{}) error {
	return s.Session.Set(s.fieldname(key, false), v)
}
func (s *Session) Load(key string, v interface{}) error {
	return s.Session.Get(s.fieldname(key, false), v)
}
func (s *Session) Remove(key string) error {
	return s.Session.Del(s.fieldname(key, false))
}
func (s *Session) IsNotFoundError(err error) bool {
	return s.IsNotFoundError(err)
}

func NewSession() *Session {
	return &Session{}
}
