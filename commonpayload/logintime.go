package commonpayload

import (
	"context"
	"strconv"
	"time"

	"github.com/herb-go/herbsecurity/authority"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/services/sessionpayload"
)

var PayloadNameLogintime = "logintime"
var logintimeBuilder = sessionpayload.BuilderFunc(func(ctx context.Context, st usersystem.SessionType, uid string, p *authority.Payloads) error {
	p.Set(PayloadNameLogintime, []byte(strconv.FormatInt(time.Now().Unix(), 10)))
	return nil
})

var LoginPayload = func(loader func(v interface{}) error) (usersystem.Directive, error) {
	return usersystem.DirectiveFunc(func(us *usersystem.UserSystem) error {
		sp, err := sessionpayload.GetService(us)
		if err != nil {
			return err
		}
		sp.AppendBuilder(logintimeBuilder)
		return nil
	}), nil
}
