package main

import (
	"fmt"
	"reflect"
	"time"

	"github.com/gogames/watchdog/main-server/pingClientManager"
	"github.com/gogames/watchdog/main-server/safeMap"
	"github.com/gogames/watchdog/main-server/store"
)

const (
	_TIME_LAYOUT        = "06-01-02 15:04"
	_MIN_PING_FREQUENCE = 1
)

var (
	_STOP_PING_CHAN = struct{}{}
	storeEngine     *store.Store
)

func initStore() {
	storeEngine = store.NewStore().
		SetStoreEngine(store.ENGINE_FILE, fmt.Sprintf(`{"serversDir":"%s","usersDir":"%s"}`, *flagServersPath, *flagUsersPath))
	go pingLoop()
	if *flagPingFrequence < _MIN_PING_FREQUENCE {
		panic(fmt.Sprintf("should be larger than %v minutes", _MIN_PING_FREQUENCE))
	}
}

var stopChanMap = safeMap.NewSafeMap()

func pingLoop() {
	for {
		select {
		case s := <-storeEngine.AddServerChan:
			c := make(chan struct{})
			if success := stopChanMap.Set(s, c); !success {
				continue
			}
			go func(server string, stopChan <-chan struct{}) {
				for {
					select {
					case tn := <-time.Tick(time.Duration(*flagPingFrequence) * time.Minute):
						pcm.Iterate(func(location string, pc pingClientManager.PingClient) {
							go func(location string, pc pingClientManager.PingClient) {
								pr, err := pc.Ping(server)
								if err != nil {
									logger.Error("can not ping server %s: %v\n", server, err)
									return
								}
								p := store.PingRet{
									Ping: fmt.Sprintf("%.3f", pr.Avg),
									Time: tn.Format(_TIME_LAYOUT),
								}
								if err = storeEngine.AppendPingRet(server, location, p); err != nil {
									logger.Critical("can not append ping result: %v\n", p)
								}
							}(location, pc)
						})
					case <-stopChan:
						stopChanMap.Delete(server)
						return
					}
				}
			}(s, c)
		case s := <-storeEngine.KickServerChan:
			if val := stopChanMap.Get(s); val != nil {
				c, ok := val.(chan struct{})
				if ok {
					logger.Debug("stop ping the server %v", s)
					c <- _STOP_PING_CHAN
				} else {
					panic(fmt.Sprintf("the value is not struct{}, but %v", reflect.TypeOf(val).Name()))
				}
			}
		}
	}
}
