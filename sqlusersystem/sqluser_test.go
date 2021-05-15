package sqlusersystem

import (
	"testing"

	"github.com/herb-go/datasource/sql/querybuilder"
	"github.com/herb-go/herbsystem"
	"github.com/herb-go/user/status"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/modules/useraccount"
	"github.com/herb-go/usersystem/modules/userpassword"
	"github.com/herb-go/usersystem/modules/userstatus"
	"github.com/herb-go/usersystem/modules/userterm"
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

func flush() {
	c := testConfig()
	database := db.New()
	err := c.Database.ApplyTo(database)
	if err != nil {
		panic(err)
	}
	_, err = database.Exec("truncate table account")
	if err != nil {
		panic(err)
	}
	_, err = database.Exec("truncate table password")
	if err != nil {
		panic(err)
	}
	_, err = database.Exec("truncate table token")
	if err != nil {
		panic(err)
	}
	_, err = database.Exec("truncate table user")
	if err != nil {
		panic(err)
	}
}
func TestService(t *testing.T) {
	InitDB()
	var err error
	s := usersystem.New()
	herbsystem.MustReady(s)
	herbsystem.MustConfigure(s)
	err = testConfig().Execute(s)
	if err != nil {
		t.Fatal(err)
	}
	sqluser := New()
	err = testConfig().ApplyToUser(sqluser)
	if err != nil {
		t.Fatal(err)
	}
	herbsystem.MustStart(s)
	herbsystem.MustStop(s)
	s = usersystem.New()
	ustatus := userstatus.MustNewAndInstallTo(s)
	upassword := userpassword.MustNewAndInstallTo(s)
	uaccounts := useraccount.MustNewAndInstallTo(s)
	uterm := userterm.MustNewAndInstallTo(s)
	herbsystem.MustReady(s)
	herbsystem.MustConfigure(s)
	err = testConfig().Execute(s)
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
	st, ok := ustatus.MustLoadStatus(uid)
	if st != status.StatusUnkown || ok {
		t.Fatal()
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
	err = upassword.UpdatePassword(uid, "ppppassword")
	if err != nil {
		t.Fatal(err)
	}
	term := uterm.MustCurrentTerm(uid)
	if term != "" {
		t.Fatal(term)
	}
	term = uterm.MustStartNewTerm(uid)
	a := uaccounts.MustAccounts(uid)
	uaccounts.MustBindAccount(uid, user.NewAccount())
	uaccounts.MustUnbindAccount(uid, user.NewAccount())
	uid = "newtestuid"
	usercreate.MustExecCreate(s, uid)

	usercreate.MustExecCreate(s, "test2")

	st, ok = ustatus.MustLoadStatus(uid)
	if st != status.StatusUnkown || !ok {
		t.Fatal(st)
	}
	ustatus.MustUpdateStatus(uid, status.StatusBanned)
	st, ok = ustatus.MustLoadStatus(uid)
	if st != status.StatusBanned || !ok {
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
	term = uterm.MustCurrentTerm(uid)

	newterm := uterm.MustStartNewTerm(uid)

	if newterm == term {
		t.Fatal(newterm)
	}
	a = uaccounts.MustAccounts(uid)
	if len(a.Data()) != 0 {
		t.Fatal()
	}
	acc := user.NewAccount()
	acc.Account = "test2"
	accid := uaccounts.MustAccountToUID(acc)
	if accid != "" {
		t.Fatal(accid)
	}
	uaccounts.MustBindAccount(uid, acc)
	accid = uaccounts.MustAccountToUID(acc)
	if accid != uid {
		t.Fatal(accid)
	}
	err = herbsystem.Catch(func() {
		uaccounts.MustBindAccount("test2", acc)
	})
	if err != user.ErrAccountBindingExists {
		t.Fatal(err)
	}
	uaccounts.MustUnbindAccount(uid, acc)

	err = herbsystem.Catch(func() {
		uaccounts.MustUnbindAccount("test2", acc)
	})
	if err != user.ErrAccountUnbindingNotExists {
		t.Fatal(err)
	}
	accid = uaccounts.MustAccountToUID(acc)
	if accid != "" {
		t.Fatal(accid, err)
	}
	uaccounts.MustBindAccount(uid, acc)

	herbsystem.MustStop(s)
	// flush()
	s = usersystem.New()
	ustatus = userstatus.MustNewAndInstallTo(s)
	upassword = userpassword.MustNewAndInstallTo(s)
	uaccounts = useraccount.MustNewAndInstallTo(s)
	uterm = userterm.MustNewAndInstallTo(s)
	herbsystem.MustReady(s)
	herbsystem.MustConfigure(s)
	err = testConfig().Execute(s)
	if err != nil {
		t.Fatal(err)
	}
	herbsystem.MustStart(s)
	defer herbsystem.MustStop(s)
	st, ok = ustatus.MustLoadStatus(uid)
	if st != status.StatusBanned || !ok {
		t.Fatal(st)
	}
	ok, err = upassword.VerifyPassword(uid, "password")
	if ok != true || err != nil {
		t.Fatal(err)
	}

	term = uterm.MustCurrentTerm(uid)

	if newterm != term {
		t.Fatal(newterm)
	}
	accid = uaccounts.MustAccountToUID(acc)
	if accid != uid {
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
		t.Fatal(users)
	}
	users = ustatus.Service.MustListUsersByStatus("", 3, false, status.StatusNormal, status.StatusBanned)
	if len(users) != 3 {
		t.Fatal(users)
	}
	users = ustatus.Service.MustListUsersByStatus("test3", 3, false, status.StatusNormal, status.StatusBanned)
	if len(users) != 2 {
		t.Fatal(users, err)
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
