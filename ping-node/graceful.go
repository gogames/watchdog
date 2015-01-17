package main

import (
	"log"
	"os"
	"syscall"

	"github.com/gogames/utils/shutdown"
)

func initShutdown() {
	functions := make([]func(), 0)

	functions = append(functions,
		func() {
			pingClient.l.Lock()
			defer pingClient.l.Unlock()
			if pingClient.pingIndicator {
				if err := pingClient.UnRegister(); err != nil {
					logger.Error("can not unregister from main server: %v\n", err)
				} else {
					logger.Debug("unregister from main server\n")
				}
			}
		},
		func() {
			logger.Close()
		},
		func() {
			log.Println("all functions are invoked")
			os.Exit(1)
		},
	)

	signals := []os.Signal{
		syscall.SIGINT,  // 2
		syscall.SIGQUIT, // 3
		syscall.SIGTERM, // 15
	}

	for _, s := range signals {
		shutdown.Register(s, functions...)
	}

	shutdown.Start()
}
