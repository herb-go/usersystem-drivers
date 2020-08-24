package memactives

import (
	"testing"
	"time"

	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/services/activesessions"
)

func testConfig() *Config {
	return &Config{
		Durations: map[usersystem.SessionType]time.Duration{
			"test": time.Second,
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
	session := usersystem.NewSession()
	err = as.OnSessionActive(session)
	if err != nil {
		t.Fatal(err)
	}
	c, err := as.Config("notexist")
	if c == nil || c.Supported != false || err != nil {
		t.Fatal(c, err)
	}
	c, err = as.Config("test")
	if c == nil || c.Supported != true || c.Duration != time.Second || err != nil {
		t.Fatal(c, err)
	}
}
