package ldapusersystem

import (
	"github.com/herb-go/user/profile"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/modules/userpassword"
	"github.com/herb-go/usersystem/modules/userprofile"
	"gopkg.in/ldap.v2"
)

type Service struct {
	LDAP          *Config
	ProfileFields []string
	ServePassword bool
}

func (s *Service) MustVerifyPassword(uid string, password string) bool {
	l, err := s.LDAP.BindUser(uid, password)
	if err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			return false
		}
		panic(err)
	}
	defer l.Close()
	return true
}
func (s *Service) MustUpdatePassword(uid string, password string) {
	err := s.LDAP.UpdatePassword(uid, password)
	if err != nil {
		panic(err)
	}
}

//PasswordChangeable return password changeable
func (s *Service) PasswordChangeable() bool {
	return true
}

//Start start service
func (s *Service) Start() error {
	return nil
}

//Stop stop service
func (s *Service) Stop() error {
	return nil
}

//Purge purge user data cache
func (s *Service) Purge(string) error {
	return nil
}
func (s *Service) MustGetProfile(id string) *profile.Profile {
	if len(s.ProfileFields) == 0 {
		return nil
	}
	l, err := s.LDAP.DialBound()
	if err != nil {
		panic(err)
	}
	defer l.Close()
	data, err := s.LDAP.search(l, id, s.ProfileFields...)
	if err != nil {
		panic(err)
	}
	profile := profile.NewProfile()
	for k := range data {
		for _, v := range data[k] {
			profile.With(k, v)
		}
	}
	return profile
}
func (s *Service) MustUpdateProfile(id string, p *profile.Profile) {
}
func (s *Service) Execute(us *usersystem.UserSystem) error {
	if s.ServePassword {
		up := userpassword.MustGetModule(us)
		if up != nil {
			up.Service = s
		}
	}

	if len(s.ProfileFields) != 0 {
		up := userprofile.MustGetModule(us)
		if up != nil {
			up.AppendService(s)
		}
	}
	return nil
}

//DirectiveFactory factory to create ldapuser directive
var DirectiveFactory = func(loader func(v interface{}) error) (usersystem.Directive, error) {
	s := &Service{}
	err := loader(s)
	if err != nil {
		return nil, err
	}
	return s, nil
}
