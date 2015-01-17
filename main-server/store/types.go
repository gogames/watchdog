package store

import (
	"encoding/json"
	"fmt"
)

type Users map[string]*User

type User struct {
	Password       string          `json:"password"`
	MonitorServers map[string]bool `json:"monitor_servers"`
}

func newUser() *User { return &User{MonitorServers: make(map[string]bool)} }

func (u User) marshal() []byte {
	b, _ := json.Marshal(u)
	return b
}

type Servers map[ServerAddr]map[Location][]PingRet

type ServerAddr string

type Location string

type PingRet struct {
	Ping string `json:"ping"`
	Time string `json:"time"`
}

func (pr PingRet) marshal() []byte {
	b, _ := json.Marshal(pr)
	return append(b, byte('\n'))
}

func (pr PingRet) String() string {
	return fmt.Sprintf("\tping: %s\ttime: %s", pr.Ping, pr.Time)
}
