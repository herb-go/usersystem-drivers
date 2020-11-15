package herbsession_test

import (
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	_ "github.com/herb-go/deprecated/cache/drivers/syncmapcache"
	_ "github.com/herb-go/deprecated/cache/marshalers/msgpackmarshaler"
	"github.com/herb-go/herb/middleware"
	"github.com/herb-go/deprecated/session"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem-drivers/herbsession"
	"github.com/herb-go/usersystem/httpusersystem/services/websession"
)

func testClientConfig() *herbsession.Config {
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
		Prefix:      "",
		StoreConfig: config,
	}
}
func TestClientService(t *testing.T) {
	var err error
	s := usersystem.New().WithKeyword("test")
	hs := websession.MustNewAndInstallTo(s)
	s.Ready()
	s.Configuring()
	err = testClientConfig().Execute(s)
	if err != nil {
		t.Fatal(err)
	}
	if hs.Service.(*herbsession.Service).Prefix != "test" {
		t.Fatal()
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
	mux.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		err = hs.Set(r, "test", r.URL.Query().Get("value"))
		if err != nil {
			panic(err)
		}
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		result := ""
		err = hs.Get(r, "test", &result)
		if err != nil {
			if hs.IsNotFoundError(err) {
				w.Write([]byte{})
				return
			}
			panic(err)
		}
		w.Write([]byte(result))
	})
	mux.HandleFunc("/del", func(w http.ResponseWriter, r *http.Request) {
		err = hs.Del(r, "test")
		if err != nil {
			panic(err)
		}
		w.Write([]byte("ok"))
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
	if session == nil || session.Type != websession.SessionType {
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

	resp, err = client.Get(server.URL + "/get")
	if err != nil {
		panic(err)
	}
	bs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if string(bs) != "" {
		t.Fatal()
	}
	resp.Body.Close()
	resp, err = client.Get(server.URL + "/set?value=testvalue")
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	resp, err = client.Get(server.URL + "/get")
	if err != nil {
		panic(err)
	}
	bs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if string(bs) != "testvalue" {
		t.Fatal()
	}
	resp.Body.Close()
	resp, err = client.Get(server.URL + "/del")
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	resp, err = client.Get(server.URL + "/get")
	if err != nil {
		panic(err)
	}
	bs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if string(bs) != "" {
		t.Fatal()
	}
	resp.Body.Close()

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
	session, err = hs.GetSession(hs.Type, "notexist")
	if err != nil {
		t.Fatal(err, sid)
	}
	if session != nil {
		t.Fatal()
	}
}

func testCacheConfig() *herbsession.Config {
	config := &session.StoreConfig{}
	config.DriverName = session.DriverNameCacheStore
	config.Cache.Driver = "syncmapcache"
	config.Cache.TTL = 3600
	config.TokenLifetime = "1h"
	config.TokenMaxLifetime = "168h"
	config.TokenContextName = "token"
	config.CookieName = "cookiename"
	config.CookiePath = "/"
	config.CookieSecure = false
	config.UpdateActiveIntervalInSecond = 100
	config.TokenLength = 32
	config.TokenPrefixMode = session.PrefixModeRaw
	return &herbsession.Config{
		Prefix:      "",
		StoreConfig: config,
	}
}
func TestCacheService(t *testing.T) {
	var err error
	s := usersystem.New().WithKeyword("test")

	hs := websession.MustNewAndInstallTo(s)
	s.Ready()
	s.Configuring()
	err = testCacheConfig().Execute(s)
	if err != nil {
		t.Fatal(err)
	}
	if hs.Service.(*herbsession.Service).Prefix != "test" {
		t.Fatal()
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
		if session == nil || session.ID == "" {
			w.Write([]byte{})
			return
		}
		w.Write([]byte(session.UID()))
	})
	mux.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		err = hs.Set(r, "test", r.URL.Query().Get("value"))
		if err != nil {
			panic(err)
		}
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		result := ""
		err = hs.Get(r, "test", &result)
		if err != nil {
			if hs.IsNotFoundError(err) {
				w.Write([]byte{})
				return
			}
			panic(err)
		}
		w.Write([]byte(result))
	})
	mux.HandleFunc("/del", func(w http.ResponseWriter, r *http.Request) {
		err = hs.Del(r, "test")
		if err != nil {
			panic(err)
		}
		w.Write([]byte("ok"))
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
	if session == nil || session.Type != websession.SessionType || session.ID == "" {
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

	resp, err = client.Get(server.URL + "/get")
	if err != nil {
		panic(err)
	}
	bs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if string(bs) != "" {
		t.Fatal()
	}
	resp.Body.Close()
	resp, err = client.Get(server.URL + "/set?value=testvalue")
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	resp, err = client.Get(server.URL + "/get")
	if err != nil {
		panic(err)
	}
	bs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if string(bs) != "testvalue" {
		t.Fatal()
	}
	resp.Body.Close()
	resp, err = client.Get(server.URL + "/del")
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	resp, err = client.Get(server.URL + "/get")
	if err != nil {
		panic(err)
	}
	bs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	if string(bs) != "" {
		t.Fatal()
	}
	resp.Body.Close()

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
	session, err = hs.GetSession(hs.Type, "notexist")
	if err != nil {
		t.Fatal(err, sid)
	}
	if session != nil {
		t.Fatal()
	}
	resp, err = client.Get(server.URL + "/login")
	if err != nil {
		panic(err)
	}
	bs, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	resp.Body.Close()
	sid = string(bs)

	ok, err := hs.RevokeSession(sid)
	if !ok || err != nil {
		t.Fatal(ok, err)
	}
	session, err = hs.GetSession(hs.Type, sid)
	if err != nil {
		t.Fatal(err, sid)
	}
	if session != nil {
		t.Fatal()
	}

}
