package store

import (
	"encoding/json"
	"fmt"
	"time"
)

type (
	Users map[string]*User
	User  struct {
		Password       string             `json:"password"`
		MonitorServers map[string]float64 `json:"monitor_servers"`
		Email          string             `json:"email"`
	}

	UserAlertInfos map[string]UserAlertInfo
	UserAlertInfo  struct {
		Email     string
		Threshold float64
	}

	Servers map[string]map[string][]PingRet

	AlertInfo struct {
		Server   string
		Location string
		Latency  float64
		Timenow  time.Time
	}
	EmailAlert struct {
		Username  string
		Email     string
		Threshold float64
		AlertInfo
		_shouldAlert bool
	}

	PingRet struct {
		Ping string `json:"ping"`
		Time string `json:"time"`
	}
)

func newUser() *User { return &User{MonitorServers: make(map[string]float64)} }

func (u User) marshal() []byte {
	b, _ := json.Marshal(u)
	return b
}

func newAlertInfo(server, location string, latency float64, tn time.Time) AlertInfo {
	return AlertInfo{
		Server:   server,
		Location: location,
		Latency:  latency,
		Timenow:  tn,
	}
}

func (e EmailAlert) shouldAlert() bool { return e._shouldAlert }

var defaultEmailAlert = EmailAlert{}

func newEmailAlert(username, email string, threshold float64, ai AlertInfo) EmailAlert {
	if threshold < 0.000001 || ai.Latency < threshold {
		return defaultEmailAlert
	}
	return EmailAlert{
		Username:  username,
		Email:     email,
		Threshold: threshold,

		AlertInfo: ai,

		_shouldAlert: true,
	}
}

func (pr PingRet) marshal() []byte {
	b, _ := json.Marshal(pr)
	return append(b, byte('\n'))
}

func (pr PingRet) String() string { return fmt.Sprintf("\tping: %s\ttime: %s", pr.Ping, pr.Time) }
