package main

import (
	"fmt"
	"net/http"
	"syscall"
	"time"

	"github.com/gogames/ping"
	"github.com/gogames/utils/signal"
	"github.com/hprose/hprose-go/hprose"
)

var (
	hproseServer = hprose.NewHttpService()
)

func initPingServer() {
	hproseServer.CrossDomainEnabled = false

	// main server invoke ping function on ping node, then collect the data
	hproseServer.AddFunction("ping", func(addr string) ping.PingResult {
		return ping.Ping(addr, 3, 10*time.Second)
	})

	// disable and run in for loop checking if main server is up
	hproseServer.AddFunction("disable", func() {
		pingClient.l.Lock()
		defer pingClient.l.Unlock()
		pingClient.pingIndicator = false
		pingClient.t = 0
	})

	go startServer()
}

func startServer() {
	for {
		func() {
			pingClient.l.Lock()
			if !pingClient.pingIndicator {
				pingClient.l.Unlock()
				logger.Info("main server is not up yet, waiting...")
				return
			}
			pingClient.l.Unlock()
			port, err := pingClient.GetServerPort()
			if err != nil {
				panic(fmt.Sprintf("can not get ping server port from main server: %v\n", err))
			}
			addr := fmt.Sprintf(":%d", port)
			if err = http.ListenAndServe(addr, hproseServer); err != nil {
				logger.Emergency("can not listen and serve: %v", err)
				if err = signal.Signal(syscall.SIGQUIT); err != nil {
					panic(fmt.Sprintf("can not send signal current process: %v\n", err))
				}
			}
		}()
		time.Sleep(time.Second)
	}
}
