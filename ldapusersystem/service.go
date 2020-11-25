package ldapusersystem

import (
	"github.com/herb-go/user/profile"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/services/userpassword"
	"github.com/herb-go/usersystem/services/userprofile"
	"gopkg.in/ldap.v2"
)

type Service struct {
	LDAP          *Config
	ProfileFields []string
	ServePassword bool
}

func (s *Service) VerifyPassword(uid string, password string) (bool, error) {
	l, err := s.LDAP.BindUser(uid, password)
	if err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			return false, nil
		}
		return false, err
	}
	defer l.Close()
	return true, nil
}
func (s *Service) UpdatePassword(uid string, password string) error {
	return s.LDAP.UpdatePassword(uid, password)
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
func (s *Service) GetProfile(id string) (*profile.Profile, error) {
	if len(s.ProfileFields) == 0 {
		return nil, nil
	}
	l, err := s.LDAP.DialBound()
	if err != nil {
		return nil, err
	}
	defer l.Close()
	data, err := s.LDAP.search(l, id, s.ProfileFields...)
	if err != nil {
		return nil, err
	}
	profile := profile.NewProfile()
	for k := range data {
		for _, v := range data[k] {
			profile.With(k, v)
		}
	}
	return profile, nil
}
func (s *Service) UpdateProfile(id string, p *profile.Profile) error {
	return nil
}
func (s *Service) Execute(us *usersystem.UserSystem) error {
	if s.ServePassword {
		up, err := userpassword.GetService(us)
		if err != nil {
			return err
		}
		if up != nil {
			up.Service = s
		}
	}

	if len(s.ProfileFields) != 0 {
		up, err := userprofile.GetService(us)
		if err != nil {
			return err
		}
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
