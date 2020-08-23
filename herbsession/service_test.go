package herbsession_test

import (
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	_ "github.com/herb-go/herb/cache/marshalers/msgpackmarshaler"
	"github.com/herb-go/herb/middleware"
	"github.com/herb-go/session"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem-drivers/herbsession"
	"github.com/herb-go/usersystem/httpusersystem/services/httpsession"
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
	config.CookieSecure = false
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
	mux := &http.ServeMux{}
	mux.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		session, err := hs.Login(r, "testuid")
		if err != nil {
			panic(err)
		}
		w.Write([]byte(session.ID))
	})
	mux.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		_, err := hs.Logout(r)
		if err != nil {
			panic(err)
		}
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/uid", func(w http.ResponseWriter, r *http.Request) {
		session, err := hs.GetRequestSession(r)
		if err != nil {
			panic(err)
		}
		if session == nil {
			w.Write([]byte{})
			return
		}
		w.Write([]byte(session.UID()))
	})
	app := middleware.New()
	app.Use(hs.SessionMiddleware())
	app.Handle(mux)
	server := httptest.NewServer(app)
	defer server.Close()
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	client := &http.Client{}
	client.Jar = jar
	resp, err := client.Get(server.URL + "/login")
	if err != nil {
		panic(err)
	}
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	sid := string(bs)
	session, err := hs.GetSession(hs.Type, sid)
	if err != nil {
		t.Fatal(err, sid)
	}
	if session == nil || session.Type != httpsession.SessionType {
		t.Fatal()
	}
	uid := session.UID()
	if uid != "testuid" {
		t.Fatal(uid)
	}
	resp, err = client.Get(server.URL + "/uid")
	if err != nil {
		panic(err)
	}
	bs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	uid = string(bs)
	if uid != "testuid" {
		t.Fatal(uid)
	}
	resp, err = client.Get(server.URL + "/logout")
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	resp, err = client.Get(server.URL + "/uid")
	if err != nil {
		panic(err)
	}
	bs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	uid = string(bs)
	if uid != "" {
		t.Fatal(uid)
	}

}
