package main

import (
	"fmt"
	"net/http"
	"syscall"

	"github.com/gogames/utils/signal"
	"github.com/hprose/hprose-go/hprose"
)

type (
	pingServerStub struct{}
)

func (pingServerStub) Ping(ctx hprose.Context) {
	ip := getIp(ctx.(*hprose.HttpContext).Request.RemoteAddr)
	pcm.Ping(getLocation(ip))
}

func (pingServerStub) GetPingInterval() int { return *flagPingInterval }

func (pingServerStub) GetServerPort() int { return *flagPingNodeServerPort }

func (pingServerStub) Register(location string, ctx hprose.Context) {
	ip := getIp(ctx.(*hprose.HttpContext).Request.RemoteAddr)
	pcm.Register(location, getUri(ip))
	if err := setLocationMapping(location, ip); err != nil {
		logger.Info(err.Error())
		panic(err)
	}
	logger.Info("%v register in\n", location)
}

func (pingServerStub) UnRegister(ctx hprose.Context) {
	ip := getIp(ctx.(*hprose.HttpContext).Request.RemoteAddr)
	pcm.UnRegister(getLocation(ip))
	logger.Info("%v unregister\n", getLocation(ip))
	deleteLocationMapping(ip)
}

var pingServer = hprose.NewHttpService()

func initPingServer() {
	pingServer.AddMethods(new(pingServerStub))
	pingServer.GetEnabled = true
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%v", *flagManagerPort), pingServer); err != nil {
			logger.Emergency("can not listen and serve ping server: %v", err)
			if err = signal.Signal(syscall.SIGQUIT); err != nil {
				panic(fmt.Errorf("can not signal the current process: %v", err))
			}
		}
	}()
}
