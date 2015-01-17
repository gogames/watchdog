package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gogames/watchdog/main-server/pingClientManager"
)

var (
	pcm             *pingClientManager.PingClientManager
	locationMapping = make(map[string]string)
	rwl             sync.RWMutex
)

func getLocation(ip string) string {
	rwl.RLock()
	defer rwl.RUnlock()
	return locationMapping[ip]
}

func setLocationMapping(location, ip string) error {
	rwl.Lock()
	defer rwl.Unlock()
	if loc, ok := locationMapping[ip]; ok {
		return fmt.Errorf("ip %v is already mapped to location %v", ip, loc)
	}
	locationMapping[ip] = location
	return nil
}

func deleteLocationMapping(ip string) {
	rwl.Lock()
	defer rwl.Unlock()
	log.Printf("delete location mapping from %v to %v\n", ip, locationMapping[ip])
	delete(locationMapping, ip)
}

func getIp(remoteAddr string) string { return strings.Split(remoteAddr, ":")[0] }

func getUri(ip string) string {
	return fmt.Sprintf("http://%s:%d/", ip, *flagPingNodeServerPort)
}

func initPingClientManager() {
	pcm = pingClientManager.NewPingClientManager(*flagPingInterval)
}
