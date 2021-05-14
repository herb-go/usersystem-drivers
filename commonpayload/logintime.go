package commonpayload

import (
	"context"
	"strconv"
	"time"

	"github.com/herb-go/herbsecurity/authority"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/modules/sessionpayload"
)

var PayloadNameLogintime = "logintime"
var logintimeBuilder = sessionpayload.BuilderFunc(func(ctx context.Context, st usersystem.SessionType, uid string, p *authority.Payloads) {
	p.Set(PayloadNameLogintime, []byte(strconv.FormatInt(time.Now().Unix(), 10)))
})

var LoginPayload = func(loader func(v interface{}) error) (usersystem.Directive, error) {
	return usersystem.DirectiveFunc(func(us *usersystem.UserSystem) error {
		sp := sessionpayload.MustGetModule(us)
		sp.AppendBuilder(logintimeBuilder)
		return nil
	}), nil
}
