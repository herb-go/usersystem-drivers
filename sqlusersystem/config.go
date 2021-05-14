package sqlusersystem

import (
	"github.com/herb-go/datasource/sql/db"
	"github.com/herb-go/datasource/sql/querybuilder"
	"github.com/herb-go/uniqueid"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/modules/useraccount"
	"github.com/herb-go/usersystem/modules/userpassword"
	"github.com/herb-go/usersystem/modules/userstatus"
	"github.com/herb-go/usersystem/modules/userterm"
)

type Config struct {
	Database      *db.Config
	TableAccount  string
	TablePassword string
	TableToken    string
	TableUser     string
	Prefix        string
}

func (c *Config) ApplyToUser(u *User) error {
	var err error

	database := db.New()
	err = c.Database.ApplyTo(database)
	if err != nil {
		return err
	}
	q := querybuilder.New()
	q.Driver = database.Driver()
	u.QueryBuilder = q
	u.DB = database
	u.UIDGenerater = uniqueid.DefaultGenerator.GenerateID
	u.Tables.AccountMapperName = c.TableAccount
	u.Tables.PasswordMapperName = c.TablePassword
	u.Tables.UserMapperName = c.TableUser
	u.Tables.TokenMapperName = c.TableToken
	u.AddTablePrefix(c.Prefix)
	return nil
}
func (c *Config) Execute(s *usersystem.UserSystem) error {
	u := New()
	err := c.ApplyToUser(u)
	if err != nil {
		return err
	}
	if c.TableUser != "" {
		ss := userstatus.MustGetModule(s)
		if ss != nil {
			ss.Service = u.User()
		}
	}
	if c.TableAccount != "" {
		ua := useraccount.MustGetModule(s)
		if ua != nil {
			ua.Service = u.Account()
		}
	}
	if c.TablePassword != "" {
		up := userpassword.MustGetModule(s)
		if up != nil {
			up.Service = u.Password()
		}
	}
	if c.TableToken != "" {
		ut := userterm.MustGetModule(s)
		if ut != nil {
			ut.Service = u.Token()
		}
	}
	return nil
}

var DirectiveFactory = func(loader func(v interface{}) error) (usersystem.Directive, error) {
	c := &Config{}
	err := loader(c)

	if err != nil {
		return nil, err
	}

	return c, nil
}
