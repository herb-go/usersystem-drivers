package redisactives

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/herb-go/datasource/redis/redispool"
	"github.com/herb-go/herbsystem"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/modules/activesessions"
	"github.com/herb-go/usersystem/usersession"
)

func newTestConfig() *Config {
	tc := &Config{}
	err := json.Unmarshal([]byte(testConfig), tc)
	if err != nil {
		panic(err)
	}
	p := redispool.New()
	err = tc.ApplyTo(p)
	if err != nil {
		panic(err)
	}
	conn := p.Open().Get()
	defer conn.Close()
	_, err = conn.Do("flushdb")
	if err != nil {
		panic(err)
	}
	return tc
}
func TestService(t *testing.T) {
	var err error
	var testDuration = 2 * time.Second
	s := usersystem.New().WithKeyword("test")
	as := activesessions.MustNewAndInstallTo(s)
	herbsystem.MustReady(s)
	herbsystem.MustConfigure(s)
	tc := newTestConfig()
	err = tc.Execute(s)
	if err != nil {
		t.Fatal(err)
	}
	herbsystem.MustStart(s)
	defer herbsystem.MustStop(s)
	as.MustOnSessionActive(nil)

	session := usersystem.NewSession()
	as.MustOnSessionActive(session)

	session.WithType("test")
	as.MustOnSessionActive(session)

	c := as.MustConfig("notexist")
	if c == nil || c.Supported != false {
		t.Fatal(c)
	}
	c = as.MustConfig("test")
	if c == nil || c.Supported != true || c.Duration != testDuration {
		t.Fatal(c)
	}

	p := usersession.MustExecInitPayloads(s, s.SystemContext(), "test", "testuid")

	session.WithPayloads(p).WithID("testid").WithType("test")
	usersession.MustExecOnSessionActive(s, session)

	a := as.MustGetActiveSessions("test", "testuid")
	if len(a) != 1 || a[0].SessionID != "testid" {
		t.Fatal(len(a))
	}
	a = as.MustGetActiveSessions("notexist", "testuid")
	if len(a) != 0 {
		t.Fatal(len(a))
	}
	time.Sleep(2 * testDuration)
	a = as.MustGetActiveSessions("test", "testuid")
	if len(a) != 0 {
		t.Fatal(len(a))
	}
	usersession.MustExecOnSessionActive(s, session)

	a = as.MustGetActiveSessions("test", "testuid")
	if len(a) != 1 || a[0].SessionID != "testid" {
		t.Fatal(len(a), err)
	}
	as.MustPurgeActiveSession(session)
	as.MustPurgeActiveSession(session)
	as.MustPurgeActiveSession(nil)

	a = as.MustGetActiveSessions("test", "testuid")
	if len(a) != 0 {
		t.Fatal(len(a))
	}
}
