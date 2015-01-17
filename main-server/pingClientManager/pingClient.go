package pingClientManager

import (
	"github.com/gogames/ping"
	"github.com/hprose/hprose-go/hprose"
)

type PingClientStub struct {
	Ping    func(string) (ping.PingResult, error)
	Disable func() error
}

type PingClient struct {
	client hprose.Client
	*PingClientStub
}

func newPingClient(uri string) PingClient {
	client := hprose.NewHttpClient(uri)
	stub := new(PingClientStub)
	client.UseService(&stub)
	return PingClient{
		client:         client,
		PingClientStub: stub,
	}
}
