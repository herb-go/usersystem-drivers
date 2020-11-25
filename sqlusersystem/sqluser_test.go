package sqlusersystem

import (
	"testing"

	"github.com/herb-go/datasource/sql/querybuilder"
	"github.com/herb-go/user/status"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem-drivers/tomluser"
	"github.com/herb-go/usersystem/services/useraccount"
	"github.com/herb-go/usersystem/services/userpassword"
	"github.com/herb-go/usersystem/services/userstatus"
	"github.com/herb-go/usersystem/services/userterm"
	"github.com/herb-go/usersystem/usercreate"
	"github.com/herb-go/usersystem/userpurge"

	"github.com/herb-go/datasource/sql/db"

	"github.com/herb-go/user"
)

func InitDB() {
	db := db.New()
	db.Init(config)
	query := querybuilder.Builder{
		Driver: config.Driver,
	}
	query.New("TRUNCATE account").MustExec(db)
	query.New("TRUNCATE password").MustExec(db)
	query.New("TRUNCATE token").MustExec(db)
	query.New("TRUNCATE user").MustExec(db)
}
func testConfig() *Config {
	c := Config{
		Database:      config,
		TableAccount:  "account",
		TablePassword: "password",
		TableToken:    "token",
		TableUser:     "user",
		Prefix:        "",
	}
	return &c
}
func TestService(t *testing.T) {
	InitDB()
	var err error
	s := usersystem.New()
	s.Ready()
	s.Configuring()
	err = testConfig().Execute(s)
	if err != nil {
		t.Fatal(err)
	}
	s.Start()
	s.Stop()
	s = usersystem.New()
	ustatus := userstatus.MustNewAndInstallTo(s)
	upassword := userpassword.MustNewAndInstallTo(s)
	uaccounts := useraccount.MustNewAndInstallTo(s)
	uterm := userterm.MustNewAndInstallTo(s)
	s.Ready()
	s.Configuring()
	err = testConfig().Execute(s)
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
	err = upassword.UpdatePassword(uid, "ppppassword")
	if err != nil {
		t.Fatal(err)
	}
	term, err := uterm.CurrentTerm(uid)
	if term != "" || err != nil {
		t.Fatal(term, err)
	}
	term, err = uterm.StartNewTerm(uid)
	if err != nil {
		t.Fatal(term, err)
	}

	a, err := uaccounts.Accounts(uid)
	if err != nil {
		t.Fatal(err)
	}
	err = uaccounts.BindAccount(uid, user.NewAccount())
	if err != nil {
		t.Fatal(err)
	}
	err = uaccounts.UnbindAccount(uid, user.NewAccount())
	if err != nil {
		t.Fatal(err)
	}
	uid = "newtestuid"
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
	if st != status.StatusUnkown {
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
	a, err = uaccounts.Accounts(uid)
	if len(a.Data()) != 0 || err != nil {
		t.Fatal(err)
	}
	acc := user.NewAccount()
	acc.Account = "test2"
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
	tomluser.Flush()
	s = usersystem.New()
	ustatus = userstatus.MustNewAndInstallTo(s)
	upassword = userpassword.MustNewAndInstallTo(s)
	uaccounts = useraccount.MustNewAndInstallTo(s)
	uterm = userterm.MustNewAndInstallTo(s)
	s.Ready()
	s.Configuring()
	err = testConfig().Execute(s)
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

	term, err = uterm.CurrentTerm(uid)
	if err != nil {
		panic(err)
	}
	if newterm != term {
		t.Fatal(newterm)
	}
	accid, err = uaccounts.AccountToUID(acc)
	if err != nil || accid != uid {
		t.Fatal(accid, err)
	}
	err = usercreate.ExecRemove(s, uid)
	if err != nil {
		t.Fatal(err)
	}
	err = ustatus.UpdateStatus("test2", status.StatusNormal)
	err = usercreate.ExecCreate(s, "test3")
	if err != nil {
		t.Fatal(err)
	}
	err = ustatus.UpdateStatus("test3", status.StatusNormal)
	err = usercreate.ExecCreate(s, "test4")
	if err != nil {
		t.Fatal(err)
	}
	err = ustatus.UpdateStatus("test4", status.StatusBanned)
	err = usercreate.ExecCreate(s, "test5")
	if err != nil {
		t.Fatal(err)
	}
	err = ustatus.UpdateStatus("test5", status.StatusBanned)
	users, err := ustatus.Service.ListUsersByStatus("", 0, status.StatusNormal, status.StatusBanned)
	if len(users) != 4 || err != nil {
		t.Fatal(users, err)
	}
	users, err = ustatus.Service.ListUsersByStatus("", 3, status.StatusNormal, status.StatusBanned)
	if len(users) != 3 || err != nil {
		t.Fatal(users, err)
	}
	users, err = ustatus.Service.ListUsersByStatus("test3", 3, status.StatusNormal, status.StatusBanned)
	if len(users) != 2 || err != nil {
		t.Fatal(users, err)
	}
	users, err = ustatus.Service.ListUsersByStatus("test2", 1, status.StatusBanned)
	if len(users) != 1 || err != nil {
		t.Fatal(users, err)
	}
	users, err = ustatus.Service.ListUsersByStatus("test2", 0, status.StatusBanned)
	if len(users) != 2 || err != nil {
		t.Fatal(users, err)
	}
}
