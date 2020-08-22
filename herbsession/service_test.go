package herbsession_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/herb-go/herb/cache/marshalers/msgpackmarshaler"
	"github.com/herb-go/herb/middleware"
	"github.com/herb-go/session"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem-drivers/herbsession"
	"github.com/herb-go/usersystem/httpusersystem/services/httpsession"
	"github.com/herb-go/usersystem/usersession"
)

func testConfig() *herbsession.Config {
	config := &session.StoreConfig{}
	config.DriverName = session.DriverNameClientStore
	config.ClientStoreKey = "test"
	config.TokenLifetime = "1h"
	config.TokenMaxLifetime = "168h"
	config.TokenContextName = "token"
	config.CookieName = "cookiename"
	config.CookiePath = "/"
	config.CookieSecure = true
	config.UpdateActiveIntervalInSecond = 100
	return &herbsession.Config{
		Prefix:      "test",
		StoreConfig: config,
	}
}
func TestService(t *testing.T) {
	var err error
	s := usersystem.New()
	ss, err := testConfig().CreateService()
	if err != nil {
		panic(err)
	}
	hs := httpsession.MustNewAndInstallTo(s)
	s.Ready()
	s.Configuring()
	err = ss.Execute(s)
	if err != nil {
		t.Fatal(err)
	}
	s.Start()
	defer s.Stop()
	app := middleware.New()
	app.Use(hs.SessionMiddleware())
	app.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := hs.GetRequestSession(r)
		if err != nil {
			panic(err)
		}
		err = usersession.Login(s, session, "test")
		if err != nil {
			panic(err)
		}
		w.Write([]byte(session.ID()))
	})
	server := httptest.NewServer(app)
	defer server.Close()
	resp, err := http.DefaultClient.Get(server.URL)
	if err != nil {
		panic(err)
	}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	sid := string(bs)
	session, err := hs.GetSession(sid, hs.Type)
	if err != nil {
		t.Fatal(err, sid)
	}
	if session == nil || session.Type() != httpsession.SessionType {
		t.Fatal()
	}
	uid, err := session.UID()
	if err != nil {
		panic(err)
	}
	if uid != "test" {
		t.Fatal(uid)
	}
}
