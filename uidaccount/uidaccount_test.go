package uidaccount

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/herb-go/herbsystem"
	"github.com/herb-go/usersystem/modules/useraccount"

	"github.com/herb-go/user"
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
	herbsystem.MustReady(s)
	herbsystem.MustConfigure(s)
	d, err := DirectiveFactory(loader.Decode)
	if err != nil {
		panic(err)
	}
	err = d.Execute(s)
	if err != nil {
		panic(err)
	}
	herbsystem.MustStart(s)
	defer herbsystem.MustStop(s)
	accounts := ua.MustAccounts("test")

	if len(accounts.Data()) != 1 || accounts.Data()[0].Keyword != testUIDAccount.AccountKeyword || accounts.Data()[0].Account != testUIDAccount.Prefix+"test"+testUIDAccount.Suffix {
		t.Fatal(ua)
	}
	uid := ua.MustAccountToUID(accounts.Data()[0])

	if uid != "test" {
		t.Fatal(s)
	}
	wrongaccount := user.NewAccount()
	wrongaccount.Keyword = testUIDAccount.AccountKeyword
	wrongaccount.Account = "test"
	err = herbsystem.Catch(func() {
		uid = ua.MustAccountToUID(wrongaccount)
	})
	if err != ErrPrefixOrSuffixNotMatch {
		t.Fatal(err)
	}
	wrongaccount = user.NewAccount()
	wrongaccount.Keyword = "wrongkeyword"
	wrongaccount.Account = testUIDAccount.Prefix + "test" + testUIDAccount.Suffix
	err = herbsystem.Catch(func() {
		uid = ua.MustAccountToUID(wrongaccount)
	})
	if err != ErrAccountKeywordNotMatch {
		t.Fatal(uid, err)
	}

	ua.MustBindAccount("test", accounts.Data()[0])

	wrongaccount = user.NewAccount()
	wrongaccount.Keyword = testUIDAccount.AccountKeyword
	wrongaccount.Account = "test"
	err = herbsystem.Catch(func() {
		ua.MustBindAccount("test", wrongaccount)
	})
	if err != ErrPrefixOrSuffixNotMatch {
		panic(err)
	}
	wrongaccount = user.NewAccount()
	wrongaccount.Keyword = "wrongkeyword"
	wrongaccount.Account = testUIDAccount.Prefix + "test" + testUIDAccount.Suffix
	err = herbsystem.Catch(func() {
		ua.MustBindAccount("test", wrongaccount)
	})
	if err != ErrAccountKeywordNotMatch {
		panic(err)
	}
	wrongaccount = user.NewAccount()
	wrongaccount.Keyword = testUIDAccount.AccountKeyword
	wrongaccount.Account = testUIDAccount.Prefix + "wrongtest" + testUIDAccount.Suffix
	err = herbsystem.Catch(func() {
		ua.MustBindAccount("test", wrongaccount)
	})
	if err != ErrUIDAndAccountNotMatch {
		panic(err)
	}

	ua.MustUnbindAccount("test", accounts.Data()[0])

	wrongaccount = user.NewAccount()
	wrongaccount.Keyword = testUIDAccount.AccountKeyword
	wrongaccount.Account = "test"
	err = herbsystem.Catch(func() {
		ua.MustUnbindAccount("test", wrongaccount)
	})
	if err != ErrPrefixOrSuffixNotMatch {
		panic(err)
	}
	wrongaccount = user.NewAccount()
	wrongaccount.Keyword = "wrongkeyword"
	wrongaccount.Account = testUIDAccount.Prefix + "test" + testUIDAccount.Suffix
	err = herbsystem.Catch(func() {
		ua.MustUnbindAccount("test", wrongaccount)
	})
	if err != ErrAccountKeywordNotMatch {
		panic(err)
	}
	wrongaccount = user.NewAccount()
	wrongaccount.Keyword = testUIDAccount.AccountKeyword
	wrongaccount.Account = testUIDAccount.Prefix + "wrongtest" + testUIDAccount.Suffix
	err = herbsystem.Catch(func() {
		ua.MustUnbindAccount("test", wrongaccount)
	})
	if err != ErrUIDAndAccountNotMatch {
		panic(err)
	}
	err = ua.Purge("test")
	if err != nil {
		panic(err)
	}
}
