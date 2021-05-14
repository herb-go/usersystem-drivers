package tomluser_test

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/herb-go/herbsecurity/authorize/role"
	"github.com/herb-go/herbsystem"
	"github.com/herb-go/user"

	"github.com/herb-go/providers/herb/statictoml"
	"github.com/herb-go/user/status"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem-drivers/tomluser"
	"github.com/herb-go/usersystem/modules/useraccount"
	"github.com/herb-go/usersystem/modules/userpassword"
	"github.com/herb-go/usersystem/modules/userprofile"
	"github.com/herb-go/usersystem/modules/userrole"
	"github.com/herb-go/usersystem/modules/userstatus"
	"github.com/herb-go/usersystem/modules/userterm"
	"github.com/herb-go/usersystem/usercreate"
	"github.com/herb-go/usersystem/userpurge"
)

func testConfig(s statictoml.Source) *tomluser.Config {
	return &tomluser.Config{
		Source:        s,
		ProfileFields: []string{"test1", "test2"},
		ServePassword: true,
		ServeStatus:   true,
		ServeAccounts: true,
		ServeRoles:    true,
		ServeTerm:     true,
		ServeProfile:  true,
	}
}
func TestService(t *testing.T) {
	var err error
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	source := statictoml.Source(path.Join(dir, "test.static.toml"))
	err = ioutil.WriteFile(string(source), []byte(""), 0700)
	if err != nil {
		t.Fatal(err)
	}
	s := usersystem.New()
	herbsystem.MustReady(s)
	herbsystem.MustConfigure(s)
	err = testConfig(source).Execute(s)
	if err != nil {
		t.Fatal(err)
	}
	herbsystem.MustStart(s)
	herbsystem.MustStop(s)

	s = usersystem.New()
	ustatus := userstatus.MustNewAndInstallTo(s)
	upassword := userpassword.MustNewAndInstallTo(s)
	uaccounts := useraccount.MustNewAndInstallTo(s)
	uprofiles := userprofile.MustNewAndInstallTo(s)
	uterm := userterm.MustNewAndInstallTo(s)
	uroles := userrole.MustNewAndInstallTo(s)
	herbsystem.MustReady(s)
	herbsystem.MustConfigure(s)
	err = testConfig(source).Execute(s)
	if err != nil {
		t.Fatal(err)
	}
	herbsystem.MustStart(s)
	uid := "test"
	ok := usercreate.MustExecExist(s, uid)
	if ok {
		t.Fatal(ok)
	}
	usercreate.MustExecCreate(s, uid)
	err = herbsystem.Catch(func() {
		usercreate.MustExecCreate(s, uid)
	})
	if err != user.ErrUserExists {
		t.Fatal(err)
	}
	ok = usercreate.MustExecExist(s, uid)
	if !ok {
		t.Fatal(ok)
	}
	usercreate.MustExecRemove(s, uid)
	err = herbsystem.Catch(func() {
		ustatus.MustLoadStatus(uid)
	})
	if err != user.ErrUserNotExists {
		t.Fatal(err)
	}
	err = herbsystem.Catch(func() {
		usercreate.MustExecRemove(s, uid)
	})
	if err != user.ErrUserNotExists {
		t.Fatal(err)
	}
	err = herbsystem.Catch(func() {
		ustatus.MustUpdateStatus(uid, status.StatusBanned)
	})
	if err != user.ErrUserNotExists {
		t.Fatal(err)
	}
	ok, err = upassword.VerifyPassword(uid, "password")
	if ok != false || err != nil {
		t.Fatal(err)
	}
	err = upassword.UpdatePassword(uid, "password")
	if err != user.ErrUserNotExists {
		t.Fatal(err)
	}
	err = herbsystem.Catch(func() {
		uroles.MustRoles(uid)
	})
	if err != user.ErrUserNotExists {
		t.Fatal(err)
	}
	err = uroles.Service.(*tomluser.Users).SetRoles(uid, nil)
	if err != user.ErrUserNotExists {
		t.Fatal(err)
	}
	err = herbsystem.Catch(func() {
		uterm.MustCurrentTerm(uid)
	})
	if err != user.ErrUserNotExists {
		t.Fatal(err)
	}
	err = herbsystem.Catch(func() {
		uterm.MustStartNewTerm(uid)
	})
	if err != user.ErrUserNotExists {
		t.Fatal(err)
	}
	err = herbsystem.Catch(func() {
		uprofiles.MustLoadProfile(uid)
	})
	if err != user.ErrUserNotExists {
		t.Fatal(err)
	}
	err = herbsystem.Catch(func() {
		uprofiles.MustUpdateProfile(nil, uid, nil)
	})
	if err != user.ErrUserNotExists {
		t.Fatal(err)
	}
	a, err := uaccounts.Accounts(uid)
	if a != nil || err != user.ErrUserNotExists {
		t.Fatal(err)
	}
	err = uaccounts.BindAccount(uid, user.NewAccount())
	if err != user.ErrUserNotExists {
		t.Fatal(err)
	}
	err = uaccounts.UnbindAccount(uid, user.NewAccount())
	if err != user.ErrUserNotExists {
		t.Fatal(err)
	}
	usercreate.MustExecCreate(s, uid)

	usercreate.MustExecCreate(s, "test2")

	st := ustatus.MustLoadStatus(uid)
	if st != status.StatusNormal {
		t.Fatal(st)
	}
	ustatus.MustUpdateStatus(uid, status.StatusBanned)

	st = ustatus.MustLoadStatus(uid)

	if st != status.StatusBanned {
		t.Fatal(st)
	}
	ok = upassword.PasswordChangeable()
	if !ok {
		t.Fatal(ok)
	}
	userpurge.MustExecPurge(s, uid)

	ok, err = upassword.VerifyPassword(uid, "password")
	if ok != false || err != nil {
		t.Fatal(err)
	}
	err = upassword.UpdatePassword(uid, "password")
	if err != nil {
		t.Fatal(err)
	}
	ok, err = upassword.VerifyPassword(uid, "password")
	if ok != true || err != nil {
		t.Fatal(err)
	}
	r := uroles.MustRoles(uid)
	if r == nil {
		t.Fatal(r)
	}
	err = uroles.Service.(*tomluser.Users).SetRoles(uid, role.NewRoles(role.NewRole("test")))
	if err != nil {
		t.Fatal(r, err)
	}
	r = uroles.MustRoles(uid)
	if !r.Contains(role.NewRoles(role.NewRole("test"))) {
		t.Fatal(r)
	}
	term := uterm.MustCurrentTerm(uid)

	newterm := uterm.MustStartNewTerm(uid)

	if newterm == term {
		t.Fatal(newterm)
	}
	p := uprofiles.MustLoadProfile(uid)
	if len(p.Data()) != 0 {
		t.Fatal(p)
	}
	p.With("test1", "test1value").With("notexist", "notexistvalue")
	uprofiles.MustUpdateProfile(nil, uid, p)

	p = uprofiles.MustLoadProfile(uid)
	if len(p.Data()) != 1 || err != nil {
		t.Fatal(p, err)
	}
	if p.Load("test1") != "test1value" || p.Load("notexist") != "" {
		t.Fatal(p)
	}
	a, err = uaccounts.Accounts(uid)
	if len(a.Data()) != 0 || err != nil {
		t.Fatal(err)
	}
	acc := user.NewAccount()
	acc.Account = "testacc"
	accid, err := uaccounts.AccountToUID(acc)
	if err != nil || accid != "" {
		t.Fatal(accid, err)
	}
	err = uaccounts.BindAccount(uid, acc)
	if err != nil {
		t.Fatal(err)
	}
	accid, err = uaccounts.AccountToUID(acc)
	if err != nil || accid != uid {
		t.Fatal(accid, err)
	}
	err = uaccounts.BindAccount("test2", acc)
	if err != user.ErrAccountBindingExists {
		t.Fatal(err)
	}
	err = uaccounts.UnbindAccount(uid, acc)
	if err != nil {
		t.Fatal(err)
	}
	err = uaccounts.UnbindAccount("test2", acc)
	if err != user.ErrAccountUnbindingNotExists {
		t.Fatal(err)
	}
	accid, err = uaccounts.AccountToUID(acc)
	if err != nil || accid != "" {
		t.Fatal(accid, err)
	}
	err = uaccounts.BindAccount(uid, acc)
	if err != nil {
		t.Fatal(err)
	}
	herbsystem.MustStop(s)
	tomluser.Flush()
	s = usersystem.New()
	ustatus = userstatus.MustNewAndInstallTo(s)
	upassword = userpassword.MustNewAndInstallTo(s)
	uaccounts = useraccount.MustNewAndInstallTo(s)
	uprofiles = userprofile.MustNewAndInstallTo(s)
	uterm = userterm.MustNewAndInstallTo(s)
	uroles = userrole.MustNewAndInstallTo(s)
	herbsystem.MustReady(s)
	herbsystem.MustConfigure(s)
	err = testConfig(source).Execute(s)
	if err != nil {
		t.Fatal(err)
	}
	herbsystem.MustStart(s)
	defer herbsystem.MustStop(s)

	st = ustatus.MustLoadStatus(uid)
	if st != status.StatusBanned {
		t.Fatal(st)
	}
	ok, err = upassword.VerifyPassword(uid, "password")
	if ok != true || err != nil {
		t.Fatal(err)
	}
	r = uroles.MustRoles(uid)
	if !r.Contains(role.NewRoles(role.NewRole("test"))) {
		t.Fatal(r)
	}

	term = uterm.MustCurrentTerm(uid)

	if newterm != term {
		t.Fatal(newterm)
	}
	p = uprofiles.MustLoadProfile(uid)
	if len(p.Data()) != 1 {
		t.Fatal(p)
	}
	if p.Load("test1") != "test1value" || p.Load("notexist") != "" {
		t.Fatal(p)
	}
	accid, err = uaccounts.AccountToUID(acc)
	if err != nil || accid != uid {
		t.Fatal(accid, err)
	}
	usercreate.MustExecRemove(s, uid)

	ustatus.MustUpdateStatus("test2", status.StatusNormal)
	usercreate.MustExecCreate(s, "test3")

	ustatus.MustUpdateStatus("test3", status.StatusNormal)
	usercreate.MustExecCreate(s, "test4")

	ustatus.MustUpdateStatus("test4", status.StatusBanned)
	usercreate.MustExecCreate(s, "test5")

	ustatus.MustUpdateStatus("test5", status.StatusBanned)
	users := ustatus.Service.MustListUsersByStatus("", 0, false, status.StatusNormal, status.StatusBanned)
	if len(users) != 4 {
		t.Fatal(users, err)
	}
	users = ustatus.Service.MustListUsersByStatus("", 3, false, status.StatusNormal, status.StatusBanned)
	if len(users) != 3 {
		t.Fatal(users)
	}
	users = ustatus.Service.MustListUsersByStatus("test3", 3, false, status.StatusNormal, status.StatusBanned)
	if len(users) != 2 {
		t.Fatal(users)
	}
	users = ustatus.Service.MustListUsersByStatus("test2", 1, false, status.StatusBanned)
	if len(users) != 1 {
		t.Fatal(users)
	}
	users = ustatus.Service.MustListUsersByStatus("test2", 0, false, status.StatusBanned)
	if len(users) != 2 {
		t.Fatal(users)
	}
}

func TestHash(t *testing.T) {
	u := tomluser.NewUser()
	u.Salt = "salt"
	p, err := tomluser.Hash("", "1234", u)
	if err != nil || p != "1234" {
		t.Fatal(p, err)
	}
	p, err = tomluser.Hash("md5", "1234", u)
	d := md5.Sum([]byte("1234" + "salt"))
	if err != nil || p != hex.EncodeToString(d[:]) {
		t.Fatal(p, err)
	}
}
