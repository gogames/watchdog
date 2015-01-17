package main

import (
	"log"
	"os"
	"syscall"

	"github.com/gogames/watchdog/main-server/pingClientManager"
	"github.com/gogames/utils/shutdown"
)

func initShutdown() {
	ss := []os.Signal{
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	}
	fs := []func(){
		func() {
			pcm.Iterate(func(_ string, pc pingClientManager.PingClient) {
				pc.Disable()
			})
		},
		func() {
			storeEngine.Close()
		},
		func() {
			sess.Close()
		},
		func() {
			logger.Close()
		},
		func() {
			log.Println("all functions are invoked")
			os.Exit(1)
		},
	}

	for _, s := range ss {
		shutdown.Register(s, fs...)
	}

	shutdown.Start()
}
