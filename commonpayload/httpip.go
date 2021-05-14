package commonpayload

import (
	"context"
	"net"

	"github.com/herb-go/usersystem/httpusersystem"

	"github.com/herb-go/herbsecurity/authority"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/modules/sessionpayload"
)

var PayloadNameHTTPIp = "httpip"
var httpipBuilder = sessionpayload.BuilderFunc(func(ctx context.Context, st usersystem.SessionType, uid string, p *authority.Payloads) {
	r := httpusersystem.GetRequest(ctx)
	if r == nil {
		return
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		panic(err)
	}
	p.Set(PayloadNameHTTPIp, []byte(ip))
})

var HTTPIpPayload = func(loader func(v interface{}) error) (usersystem.Directive, error) {
	return usersystem.DirectiveFunc(func(us *usersystem.UserSystem) error {
		sp := sessionpayload.MustGetModule(us)
		sp.AppendBuilder(httpipBuilder)
		return nil
	}), nil
}
