package uidaccount

import (
	"errors"
	"strings"

	"github.com/herb-go/user"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/modules/useraccount"
)

// ErrUIDAndAccountNotMatch error raised when uid and account not match
var ErrUIDAndAccountNotMatch = errors.New("uidaccount:uid and account not match")

// ErrAccountKeywordNotMatch error raised when account keyword not match
var ErrAccountKeywordNotMatch = errors.New("uidaccount: account keyword not match")

//ErrPrefixOrSuffixNotMatch error raised when prefix or suffix not match
var ErrPrefixOrSuffixNotMatch = errors.New("uidaccount:prefix or suffix not match")

//UIDAccount uidaccount directive struct
type UIDAccount struct {
	//AccountKeyword account keyword
	AccountKeyword string
	//Prefix uid prefix
	Prefix string
	//Suffix uid suffix
	Suffix string
}

func (u *UIDAccount) accountToUID(account string) (string, error) {
	if strings.HasPrefix(account, u.Prefix) && strings.HasSuffix(account, u.Suffix) {
		return strings.TrimSuffix(strings.TrimPrefix(account, u.Prefix), u.Suffix), nil
	}
	return "", ErrPrefixOrSuffixNotMatch
}
func (u *UIDAccount) uidToAccount(uid string) string {
	return u.Prefix + uid + u.Suffix
}

//Accounts return account of give uid.
func (u *UIDAccount) MustAccounts(uid string) *user.Accounts {
	account := user.NewAccount()
	account.Keyword = u.AccountKeyword
	account.Account = u.uidToAccount(uid)
	return &user.Accounts{account}
}

//AccountToUID query uid by user account.
//Return user id and any error if raised.
//Return empty string as userid if account not found.
func (u *UIDAccount) MustAccountToUID(account *user.Account) string {
	if account.Keyword != u.AccountKeyword {
		panic(ErrAccountKeywordNotMatch)
	}
	uid, err := u.accountToUID(account.Account)
	if err != nil {
		panic(err)
	}
	return uid
}

//BindAccount bind account to user.
//If account exists,user.ErrAccountBindingExists should be rasied.
func (u *UIDAccount) MustBindAccount(uid string, account *user.Account) {
	if account.Keyword != u.AccountKeyword {
		panic(ErrAccountKeywordNotMatch)
	}
	id, err := u.accountToUID(account.Account)
	if err != nil {
		panic(err)
	}
	if uid != id {
		panic(ErrUIDAndAccountNotMatch)
	}
}

//UnbindAccount unbind account from user.
//Return any error if raised.
//If account not exists,user.ErrAccountUnbindingNotExists should be rasied.
func (u *UIDAccount) MustUnbindAccount(uid string, account *user.Account) {
	if account.Keyword != u.AccountKeyword {
		panic(ErrAccountKeywordNotMatch)
	}
	id, err := u.accountToUID(account.Account)
	if err != nil {
		panic(err)
	}
	if uid != id {
		panic(ErrUIDAndAccountNotMatch)
	}
}

//Purge purge user data cache
func (u *UIDAccount) Purge(string) error {
	return nil
}

//Start start service
func (u *UIDAccount) Start() error {
	return nil
}

//Stop stop service
func (u *UIDAccount) Stop() error {
	return nil
}

//Execute apply uidaccount directive to usersystem
func (u *UIDAccount) Execute(s *usersystem.UserSystem) error {
	ua := useraccount.MustGetModule(s)

	if ua == nil {
		return nil
	}
	ua.Service = u
	return nil
}

//DirectiveFactory factory to create uidaccount directive
var DirectiveFactory = func(loader func(v interface{}) error) (usersystem.Directive, error) {
	c := &UIDAccount{}
	err := loader(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
