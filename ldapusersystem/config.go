package ldapusersystem

import (
	"fmt"

	"github.com/herb-go/user"

	"gopkg.in/ldap.v2"
)

//Config ldap user config struct
// example:
// Net:          "tcp",
// Addr:         "127.0.0.1:389",
// UserPattern:  "uid=%s,ou=People,dc=example",
// BindDN:       "cn=admin,dc=example",
// BindPass:     "password",
// SearchDN:     "ou=People,dc=example",
// SearchFilter: "(uid=%s)",
// GroupDN:      "ou=Group,dc=example",
// GroupFilter:  "(member=%s)",
// GroupIDField: "cn",
type Config struct {
	Net          string
	Addr         string
	UserPattern  string
	BindDN       string
	BindPass     string
	SearchDN     string
	SearchFilter string
	GroupDN      string
	GroupIDField string
	GroupFilter  string
}

//UpdatePassword update user password
//Return any error if raised
func (c *Config) UpdatePassword(uid string, password string) error {
	uid = ldap.EscapeFilter(uid)
	l, err := c.Dial()
	if err != nil {
		return err
	}
	defer l.Close()

	err = l.Bind(c.BindDN, c.BindPass)
	if err != nil {
		return err
	}
	passwordModifyRequest := ldap.NewPasswordModifyRequest(fmt.Sprintf(c.UserPattern, uid), "", password)
	_, err = l.PasswordModify(passwordModifyRequest)
	return err
}
func (c *Config) search(l *ldap.Conn, id string, fields ...string) (map[string][]string, error) {
	id = ldap.EscapeFilter(id)
	searchRequest := ldap.NewSearchRequest(
		c.SearchDN,
		ldap.ScopeSingleLevel,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf(c.SearchFilter, id),
		fields,
		nil)
	result, err := l.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	if len(result.Entries) == 0 {
		err = user.ErrUserNotExists
		return nil, err
	}
	data := map[string][]string{}
	for _, v := range fields {
		data[v] = result.Entries[0].GetAttributeValues(v)
	}
	return data, nil
}
func (c *Config) SearchUser(id string, fields ...string) (map[string][]string, error) {
	l, err := c.DialBound()
	if err != nil {
		return nil, err
	}
	defer l.Close()
	return c.search(l, id, fields...)
}

func (c *Config) SearchUserGroups(id string) ([]string, error) {
	id = ldap.EscapeFilter(id)

	l, err := c.DialBound()
	if err != nil {
		return nil, err
	}
	uid := fmt.Sprintf(c.UserPattern, id)
	searchRequest := ldap.NewSearchRequest(
		c.GroupDN,
		ldap.ScopeSingleLevel,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf(c.GroupFilter, uid),
		[]string{c.GroupIDField},
		nil)
	result, err := l.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	var data []string
	resultLen := len(result.Entries)
	if resultLen > 0 {
		data = make([]string, resultLen)
		for i := 0; i < resultLen; i++ {
			data[i] = result.Entries[i].GetAttributeValue(c.GroupIDField)
		}
	}
	return data, nil
}

func (c *Config) Dial() (*ldap.Conn, error) {
	return ldap.Dial(c.Net, c.Addr)
}

func (c *Config) DialBound() (*ldap.Conn, error) {
	l, err := c.Dial()
	if err != nil {
		return nil, err
	}
	err = l.Bind(c.BindDN, c.BindPass)
	if err != nil {
		return nil, err
	}
	return l, nil
}
func (c *Config) BindUser(uid, password string) (*ldap.Conn, error) {
	uid = ldap.EscapeFilter(uid)
	l, err := c.Dial()
	if err != nil {
		return nil, err
	}
	err = l.Bind(fmt.Sprintf(c.UserPattern, uid), password)
	return l, err
}
