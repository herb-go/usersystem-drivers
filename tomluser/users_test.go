package tomluser_test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/herb-go/herb/user"

	"github.com/herb-go/herb/user/status"
	"github.com/herb-go/providers/herb/statictoml"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem-drivers/tomluser"
	"github.com/herb-go/usersystem/services/userpassword"
	"github.com/herb-go/usersystem/services/userstatus"
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

	// uaccounts := useraccount.MustNewAndInstallTo(s)
	// uprofiles := userprofile.MustNewAndInstallTo(s)
	// uterm := userterm.MustNewAndInstallTo(s)
	// uroles := userrole.MustNewAndInstallTo(s)
	s.Ready()
	s.Configuring()
	err = testConfig(source).Execute(s)
	if err != nil {
		t.Fatal(err)
	}
	s.Start()
	defer s.Stop()
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
	err = usercreate.ExecCreate(s, uid)
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

}
