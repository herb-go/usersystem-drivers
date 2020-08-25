package memactives

import (
	"testing"
	"time"

	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/services/activesessions"
	"github.com/herb-go/usersystem/usersession"
)

var testDuration = time.Millisecond

func testConfig() *Config {
	return &Config{
		Durations: map[string]string{
			"test": testDuration.String(),
		},
	}
}

func TestService(t *testing.T) {
	var err error
	s := usersystem.New().WithKeyword("test")
	as := activesessions.MustNewAndInstallTo(s)
	s.Ready()
	s.Configuring()
	err = testConfig().Execute(s)
	if err != nil {
		t.Fatal(err)
	}
	s.Start()
	defer s.Stop()
	err = as.OnSessionActive(nil)
	if err != nil {
		t.Fatal(err)
	}
	session := usersystem.NewSession()
	err = as.OnSessionActive(session)
	if err != nil {
		t.Fatal(err)
	}
	session.WithType("test")
	err = as.OnSessionActive(session)
	if err != nil {
		t.Fatal(err)
	}
	c, err := as.Config("notexist")
	if c == nil || c.Supported != false || err != nil {
		t.Fatal(c, err)
	}
	c, err = as.Config("test")
	if c == nil || c.Supported != true || c.Duration != testDuration || err != nil {
		t.Fatal(c, err)
	}

	p, err := usersession.ExecInitPayloads(s, s.Context, "test", "testuid")
	if err != nil {
		t.Fatal(err)
	}
	session.WithPayloads(p).WithID("testid").WithType("test")
	err = usersession.ExecOnSessionActive(s, session)
	if err != nil {
		t.Fatal(err)
	}
	a, err := as.GetActiveSessions("test", "testuid")
	if err != nil || len(a) != 1 || a[0].SessionID != "testid" {
		t.Fatal(len(a), err)
	}
	a, err = as.GetActiveSessions("notexist", "testuid")
	if err != nil || len(a) != 0 {
		t.Fatal(len(a), err)
	}
	service := as.Service.(*Service)
	service.Stores["test"].Update()
	time.Sleep(2 * testDuration)
	a, err = as.GetActiveSessions("test", "testuid")
	if err != nil || len(a) != 0 {
		t.Fatal(len(a), err)
	}
	err = usersession.ExecOnSessionActive(s, session)
	if err != nil {
		t.Fatal(err)
	}
	a, err = as.GetActiveSessions("test", "testuid")
	if err != nil || len(a) != 1 || a[0].SessionID != "testid" {
		t.Fatal(len(a), err)
	}
	err = as.PurgeActiveSession("", "testuid", session.Payloads.LoadString(activesessions.PayloadSerialNumber))
	if err != nil {
		t.Fatal()
	}
	err = as.PurgeActiveSession("test", "testuid", session.Payloads.LoadString(activesessions.PayloadSerialNumber))
	if err != nil {
		t.Fatal()
	}
	err = as.PurgeActiveSession("test", "", "")
	if err != nil {
		t.Fatal()
	}

	a, err = as.GetActiveSessions("test", "testuid")
	if err != nil || len(a) != 0 {
		t.Fatal(len(a), err)
	}
}
