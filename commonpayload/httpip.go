package commonpayload

import (
	"context"
	"net"

	"github.com/herb-go/usersystem/httpusersystem"

	"github.com/herb-go/herbsecurity/authority"
	"github.com/herb-go/usersystem"
	"github.com/herb-go/usersystem/services/sessionpayload"
)

var PayloadNameHTTPIp = "httpip"
var httpipBuilder = sessionpayload.BuilderFunc(func(ctx context.Context, st usersystem.SessionType, uid string, p *authority.Payloads) error {
	r := httpusersystem.GetRequest(ctx)
	if r == nil {
		return nil
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return err
	}
	p.Set(PayloadNameHTTPIp, []byte(ip))
	return nil
})

var HTTPIpPayload = func(loader func(v interface{}) error) (usersystem.Directive, error) {
	return usersystem.DirectiveFunc(func(us *usersystem.UserSystem) error {
		sp, err := sessionpayload.GetService(us)
		if err != nil {
			return err
		}
		sp.AppendBuilder(httpipBuilder)
		return nil
	}), nil
}
