package pingClientManager

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrorExistedLocation = errors.New("the location is already existed, can not register again")
)

type PingClientManager struct {
	pingClients  map[string]PingClient
	lastPingTime map[string]time.Time

	pingInterval time.Duration
	rwl          sync.RWMutex
}

func NewPingClientManager(pingInterval int) *PingClientManager {
	pcm := &PingClientManager{
		pingClients:  make(map[string]PingClient),
		lastPingTime: make(map[string]time.Time),
		pingInterval: time.Duration(pingInterval) * time.Second,
	}
	go pcm.start()
	return pcm
}

func (pcm *PingClientManager) Register(location, uri string) (err error) {
	err = ErrorExistedLocation
	pcm.withWriteLock(func() {
		if _, ok := pcm.pingClients[location]; !ok {
			pcm.pingClients[location] = newPingClient(uri)
			pcm.lastPingTime[location] = time.Now()
		}
	})
	return nil
}

func (pcm *PingClientManager) UnRegister(location string) {
	pcm.withWriteLock(func() { pcm.kick(location) })
}

func (pcm *PingClientManager) Iterate(f func(location string, pc PingClient)) {
	pcm.withReadLock(func() {
		for location, pc := range pcm.pingClients {
			f(location, pc)
		}
	})
}

func (pcm *PingClientManager) Ping(location string) {
	pcm.withReadLock(func() {
		if _, ok := pcm.lastPingTime[location]; ok {
			pcm.lastPingTime[location] = time.Now()
		}
	})
}

func (pcm *PingClientManager) withWriteLock(f func()) {
	pcm.rwl.Lock()
	defer pcm.rwl.Unlock()
	f()
}

func (pcm *PingClientManager) withReadLock(f func()) {
	pcm.rwl.RLock()
	defer pcm.rwl.RUnlock()
	f()
}

func (pcm *PingClientManager) start() {
	f := func(tn time.Time) {
		for location, t := range pcm.lastPingTime {
			if tn.Sub(t) > pcm.pingInterval {
				pcm.kick(location)
			}
		}
	}
	for {
		select {
		case t := <-time.Tick(pcm.pingInterval):
			pcm.withWriteLock(func() { f(t) })
		}
	}
}

func (pcm *PingClientManager) kick(location string) {
	fmt.Println("kick", location)
	delete(pcm.pingClients, location)
	delete(pcm.lastPingTime, location)
}
