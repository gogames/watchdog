package main

import (
	"fmt"
	"reflect"
	"time"

	"github.com/gogames/utils/sendmail"
	"github.com/gogames/watchdog/main-server/pingClientManager"
	"github.com/gogames/watchdog/main-server/safeMap"
	"github.com/gogames/watchdog/main-server/store"
)

const (
	_MIN_PING_FREQUENCE = 1
	_TIME_LAYOUT        = "2006-01-02 15:04"
	_FROM               = "watchdog"
	_SUBJECT            = "Alert"
	_BODY               = "Dear %s,\n\tPing latency from %s to the server %s you are monitoring is %v on %v.\n\n\nThis is an alert automatically sent by http://watchdog.top/, please do not reply!\n\n"
)

var (
	_STOP_PING_CHAN = struct{}{}
	storeEngine     *store.Store
)

func initStore() {
	storeEngine = store.NewStore().
		SetStoreEngine(store.ENGINE_FILE, fmt.Sprintf(`{"serversDir":"%s","usersDir":"%s"}`, *flagServersPath, *flagUsersPath))
	if *flagPingFrequence < _MIN_PING_FREQUENCE {
		panic(fmt.Sprintf("should be larger than %v minutes", _MIN_PING_FREQUENCE))
	}
	go pingLoop()
	go alertLoop()
}

var stopChanMap = safeMap.NewSafeMap()

func pingLoop() {
	var iterator = func(server, location string, pc pingClientManager.PingClient, tn time.Time) {
		go func(location string, pc pingClientManager.PingClient) {
			pr, err := pc.Ping(server)
			if err != nil {
				logger.Error("can not ping server %s: %v\n", server, err)
				return
			}
			if err = storeEngine.AppendPingRet(server, location, pr.Avg, tn); err != nil {
				logger.Critical("can not append ping result of %v: %v\n", server, err)
			}
		}(location, pc)
	}

	var ti = time.Duration(*flagPingFrequence) * time.Minute
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
					case tn := <-time.Tick(ti):
						pcm.Iterate(
							func(loc string, pc pingClientManager.PingClient) {
								iterator(server, loc, pc, tn)
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

func alertLoop() {
	for {
		ea := <-storeEngine.EmailALertChan
		go sendmail.Send(_FROM,
			ea.Email,
			_SUBJECT,
			fmt.Sprintf(_BODY, ea.Username, ea.Location, ea.Server, ea.Latency, ea.Timenow.Format(_TIME_LAYOUT)))
	}
}
