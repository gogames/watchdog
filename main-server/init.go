package main

import "runtime"

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	initFlag()
	initLogger()
	initShutdown()
	initPingServer()
	initPingClientManager()
	initSession()
	initMainServer()
	initStore()
}
