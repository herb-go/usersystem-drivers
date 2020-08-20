package tomluser_test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/herb-go/herb/user"
	"github.com/herb-go/herbsecurity/authorize/role"

	"github.com/herb-go/herb/user/status"
	"github.com/herb-go/providers/herb/statictoml"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem-drivers/tomluser"
	"github.com/herb-go/usersystem/services/useraccount"
	"github.com/herb-go/usersystem/services/userpassword"
	"github.com/herb-go/usersystem/services/userprofile"
	"github.com/herb-go/usersystem/services/userrole"
	"github.com/herb-go/usersystem/services/userstatus"
	"github.com/herb-go/usersystem/services/userterm"
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
	s.Ready()
	s.Configuring()
	err = testConfig(source).Execute(s)
	if err != nil {
		t.Fatal(err)
	}
	s.Start()
	s.Stop()
	s = usersystem.New()
	ustatus := userstatus.MustNewAndInstallTo(s)
	upassword := userpassword.MustNewAndInstallTo(s)
	uaccounts := useraccount.MustNewAndInstallTo(s)
	uprofiles := userprofile.MustNewAndInstallTo(s)
	uterm := userterm.MustNewAndInstallTo(s)
	uroles := userrole.MustNewAndInstallTo(s)
	s.Ready()
	s.Configuring()
	err = testConfig(source).Execute(s)
	if err != nil {
		t.Fatal(err)
	}
	s.Start()
	uid := "test"
	ok, err := usercreate.ExecExist(s, uid)
	if ok || err != nil {
		t.Fatal(ok, err)
	}
	err = usercreate.ExecCreate(s, uid)
	if err != nil {
		t.Fatal(err)
	}
	err = usercreate.ExecCreate(s, uid)
	if err != user.ErrUserExists {
		t.Fatal(err)
	}
	ok, err = usercreate.ExecExist(s, uid)
	if !ok || err != nil {
		t.Fatal(ok, err)
	}
	err = usercreate.ExecRemove(s, uid)
	if err != nil {
		t.Fatal(err)
	}
	st, err := ustatus.LoadStatus(uid)
	if err != user.ErrUserNotExists {
		t.Fatal(err)
	}
	err = usercreate.ExecRemove(s, uid)
	if err != user.ErrUserNotExists {
		t.Fatal(err)
	}
	err = ustatus.UpdateStatus(uid, status.StatusBanned)
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
	r, err := uroles.Roles(uid)
	if r != nil || err != user.ErrUserNotExists {
		t.Fatal(r, err)
	}
	err = uroles.Service.(*tomluser.Users).SetRoles(uid, nil)
	if err != user.ErrUserNotExists {
		t.Fatal(r, err)
	}
	term, err := uterm.CurrentTerm(uid)
	if term != "" || err != user.ErrUserNotExists {
		t.Fatal(r, err)
	}
	term, err = uterm.StartNewTerm(uid)
	if term != "" || err != user.ErrUserNotExists {
		t.Fatal(r, err)
	}
	p, err := uprofiles.LoadProfile(uid)
	if p != nil || err != user.ErrUserNotExists {
		t.Fatal(p, err)
	}
	err = uprofiles.UpdateProfile(nil, uid, p)
	if err != user.ErrUserNotExists {
		t.Fatal(err)
	}
	a, err := uaccounts.Account(uid)
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
	err = usercreate.ExecCreate(s, uid)
	if err != nil {
		t.Fatal(err)
	}
	err = usercreate.ExecCreate(s, "test2")
	if err != nil {
		t.Fatal(err)
	}
	st, err = ustatus.LoadStatus(uid)
	if err != nil {
		t.Fatal(err)
	}
	if st != status.StatusNormal {
		t.Fatal(st)
	}
	err = ustatus.UpdateStatus(uid, status.StatusBanned)
	if err != nil {
		t.Fatal(err)
	}
	st, err = ustatus.LoadStatus(uid)
	if err != nil {
		t.Fatal(err)
	}
	if st != status.StatusBanned {
		t.Fatal(st)
	}
	ok = upassword.PasswordChangeable()
	if !ok {
		t.Fatal(ok)
	}
	err = userpurge.ExecPurge(s, uid)
	if err != nil {
		t.Fatal(err)
	}
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
	r, err = uroles.Roles(uid)
	if r == nil || err != nil {
		t.Fatal(r, err)
	}
	err = uroles.Service.(*tomluser.Users).SetRoles(uid, role.NewRoles(role.NewRole("test")))
	if err != nil {
		t.Fatal(r, err)
	}
	r, err = uroles.Roles(uid)
	if !r.Contains(role.NewRoles(role.NewRole("test"))) || err != nil {
		t.Fatal(r, err)
	}
	term, err = uterm.CurrentTerm(uid)
	if err != nil {
		panic(err)
	}
	newterm, err := uterm.StartNewTerm(uid)
	if err != nil {
		panic(err)
	}
	if newterm == term {
		t.Fatal(newterm)
	}
	p, err = uprofiles.LoadProfile(uid)
	if len(p.Data()) != 0 || err != nil {
		t.Fatal(p, err)
	}
	p.With("test1", "test1value").With("notexist", "notexistvalue")
	err = uprofiles.UpdateProfile(nil, uid, p)
	if err != nil {
		t.Fatal(p, err)
	}
	p, err = uprofiles.LoadProfile(uid)
	if len(p.Data()) != 1 || err != nil {
		t.Fatal(p, err)
	}
	if p.Load("test1") != "test1value" || p.Load("notexist") != "" {
		t.Fatal(p)
	}
	a, err = uaccounts.Account(uid)
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
	s.Stop()
	s = usersystem.New()
	ustatus = userstatus.MustNewAndInstallTo(s)
	upassword = userpassword.MustNewAndInstallTo(s)
	uaccounts = useraccount.MustNewAndInstallTo(s)
	uprofiles = userprofile.MustNewAndInstallTo(s)
	uterm = userterm.MustNewAndInstallTo(s)
	uroles = userrole.MustNewAndInstallTo(s)
	s.Ready()
	s.Configuring()
	err = testConfig(source).Execute(s)
	if err != nil {
		t.Fatal(err)
	}
	s.Start()
	defer s.Stop()
	st, err = ustatus.LoadStatus(uid)
	if err != nil {
		t.Fatal(err)
	}
	if st != status.StatusBanned {
		t.Fatal(st)
	}
	ok, err = upassword.VerifyPassword(uid, "password")
	if ok != true || err != nil {
		t.Fatal(err)
	}
	r, err = uroles.Roles(uid)
	if !r.Contains(role.NewRoles(role.NewRole("test"))) || err != nil {
		t.Fatal(r, err)
	}

	term, err = uterm.CurrentTerm(uid)
	if err != nil {
		panic(err)
	}
	if newterm != term {
		t.Fatal(newterm)
	}
	p, err = uprofiles.LoadProfile(uid)
	if len(p.Data()) != 1 || err != nil {
		t.Fatal(p, err)
	}
	if p.Load("test1") != "test1value" || p.Load("notexist") != "" {
		t.Fatal(p)
	}
	accid, err = uaccounts.AccountToUID(acc)
	if err != nil || accid != uid {
		t.Fatal(accid, err)
	}
}
