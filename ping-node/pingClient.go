package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/hprose/hprose-go/hprose"
)

const _MAX_PING_ERROR = 5

// invoke functions provided by main server
type pingClientStub struct {
	Ping            func() error
	GetPingInterval func() (int, error)
	GetServerPort   func() (int, error)
	Register        func(location string) error // location of the ping node
	UnRegister      func() error
}

type pingClientStruct struct {
	pingClientStub

	interval time.Duration

	l             sync.Mutex
	pingIndicator bool
	t             int64
}

var (
	pingClient   = new(pingClientStruct)
	hproseClient hprose.Client
)

func enable() {
	for {
		func() {
			pingClient.l.Lock()
			defer pingClient.l.Unlock()
			if !pingClient.pingIndicator {
				pingClient.pingIndicator = func() bool {
					if err := pingClient.Register(*flagLocation); err != nil {
						logger.Debug(fmt.Sprintf("can not register ping node: %v\n", err))
						return false
					}

					if i, err := pingClient.GetPingInterval(); err != nil {
						logger.Debug(fmt.Sprintf("can not get ping interval: %v\n", err))
						return false
					} else {
						pingClient.interval = time.Duration(i) * time.Second / 2
					}
					return true
				}()
				if pingClient.pingIndicator {
					logger.Debug("the server is up, ping interval is %v", pingClient.interval)
				}
			}
		}()
		time.Sleep(5 * time.Second)
	}
}

func pingLoop() {
	for {
		select {
		case <-time.Tick(pingClient.interval):
			func() {
				pingClient.l.Lock()
				defer pingClient.l.Unlock()
				if !pingClient.pingIndicator {
					return
				}
				if err := pingClient.Ping(); err != nil {
					logger.Debug("can not ping main server: %v\n", err)
					pingClient.t++
					if pingClient.t >= _MAX_PING_ERROR {
						pingClient.pingIndicator = false
						pingClient.t = 0
					}
				} else {
					// reset
					pingClient.t = 0
				}
			}()
		}
	}
}

func initPingClient() {
	hproseClient = hprose.NewHttpClient("http://" + *flagMainServerAddress)
	hproseClient.UseService(&pingClient.pingClientStub)

	pingClient.interval = time.Second
	go enable()
	go pingLoop()
}
