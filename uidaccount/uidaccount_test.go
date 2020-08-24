package uidaccount

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/herb-go/usersystem/services/useraccount"

	"github.com/herb-go/herb/user"
	"github.com/herb-go/usersystem"
)

func TestUIDAccount(t *testing.T) {
	var testUIDAccount = &UIDAccount{
		AccountKeyword: "testkeyword",
		Prefix:         "testprefix",
		Suffix:         "testsuffix",
	}
	data, err := json.Marshal(testUIDAccount)
	if err != nil {
		panic(err)
	}
	loader := json.NewDecoder(bytes.NewBuffer(data))

	s := usersystem.New()
	ua := useraccount.MustNewAndInstallTo(s)
	s.Ready()
	s.Configuring()
	d, err := DirectiveFactory(loader.Decode)
	if err != nil {
		panic(err)
	}
	err = d.Execute(s)
	if err != nil {
		panic(err)
	}
	s.Start()
	defer s.Stop()
	accounts, err := ua.Account("test")
	if err != nil {
		panic(err)
	}
	if len(accounts.Data()) != 1 || accounts.Data()[0].Keyword != testUIDAccount.AccountKeyword || accounts.Data()[0].Account != testUIDAccount.Prefix+"test"+testUIDAccount.Suffix {
		t.Fatal(ua)
	}
	uid, err := ua.AccountToUID(accounts.Data()[0])
	if err != nil {
		panic(err)
	}
	if uid != "test" {
		t.Fatal(s)
	}
	wrongaccount := user.NewAccount()
	wrongaccount.Keyword = testUIDAccount.AccountKeyword
	wrongaccount.Account = "test"
	uid, err = ua.AccountToUID(wrongaccount)
	if err != ErrPrefixOrSuffixNotMatch || uid != "" {
		t.Fatal(uid, err)
	}
	wrongaccount = user.NewAccount()
	wrongaccount.Keyword = "wrongkeyword"
	wrongaccount.Account = testUIDAccount.Prefix + "test" + testUIDAccount.Suffix
	uid, err = ua.AccountToUID(wrongaccount)
	if err != ErrAccountKeywordNotMatch || uid != "" {
		t.Fatal(uid, err)
	}

	err = ua.BindAccount("test", accounts.Data()[0])
	if err != nil {
		panic(err)
	}
	wrongaccount = user.NewAccount()
	wrongaccount.Keyword = testUIDAccount.AccountKeyword
	wrongaccount.Account = "test"
	err = ua.BindAccount("test", wrongaccount)
	if err != ErrPrefixOrSuffixNotMatch {
		panic(err)
	}
	wrongaccount = user.NewAccount()
	wrongaccount.Keyword = "wrongkeyword"
	wrongaccount.Account = testUIDAccount.Prefix + "test" + testUIDAccount.Suffix
	err = ua.BindAccount("test", wrongaccount)
	if err != ErrAccountKeywordNotMatch {
		panic(err)
	}
	wrongaccount = user.NewAccount()
	wrongaccount.Keyword = testUIDAccount.AccountKeyword
	wrongaccount.Account = testUIDAccount.Prefix + "wrongtest" + testUIDAccount.Suffix
	err = ua.BindAccount("test", wrongaccount)
	if err != ErrUIDAndAccountNotMatch {
		panic(err)
	}

	err = ua.UnbindAccount("test", accounts.Data()[0])
	if err != nil {
		panic(err)
	}

	wrongaccount = user.NewAccount()
	wrongaccount.Keyword = testUIDAccount.AccountKeyword
	wrongaccount.Account = "test"
	err = ua.UnbindAccount("test", wrongaccount)
	if err != ErrPrefixOrSuffixNotMatch {
		panic(err)
	}
	wrongaccount = user.NewAccount()
	wrongaccount.Keyword = "wrongkeyword"
	wrongaccount.Account = testUIDAccount.Prefix + "test" + testUIDAccount.Suffix
	err = ua.UnbindAccount("test", wrongaccount)
	if err != ErrAccountKeywordNotMatch {
		panic(err)
	}
	wrongaccount = user.NewAccount()
	wrongaccount.Keyword = testUIDAccount.AccountKeyword
	wrongaccount.Account = testUIDAccount.Prefix + "wrongtest" + testUIDAccount.Suffix
	err = ua.UnbindAccount("test", wrongaccount)
	if err != ErrUIDAndAccountNotMatch {
		panic(err)
	}
	err = ua.Purge("test")
	if err != nil {
		panic(err)
	}
}
