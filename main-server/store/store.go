package store

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	_ERROR_INCORRECT_PASSWORD = errors.New("incorrect password")
	engines                   = make(map[string]func() StoreEngine)
)

const (
	_MIN_LEN_SERVER_CHAN = 1 << 10
)

func Register(engineName string, f func() StoreEngine) error {
	if _, ok := engines[engineName]; ok {
		return fmt.Errorf("engine %v already exist", engineName)
	}
	engines[engineName] = f
	return nil
}

type StoreEngine interface {
	LoadConfig(config string)
	Init() (Servers, Users, map[string]int64)

	WriteUser(username string, u *User) error
	AppendPingRet(server, location string, pr PingRet) error
	BatchWritePingRets(server, location string, prs []PingRet) error
}

type Store struct {
	servers    Servers
	users      Users
	allServers map[string]int64
	rwl        sync.RWMutex

	storeEngine StoreEngine

	AddServerChan  chan string
	KickServerChan chan string

	closeCounter *int64
	isClosed     bool
}

func NewStore() *Store { return &Store{closeCounter: new(int64)} }

func (s *Store) SetStoreEngine(engineName string, config string) *Store {
	if f, ok := engines[engineName]; !ok {
		panic(fmt.Errorf("store engine %v does not exist", engineName))
	} else {
		s.storeEngine = f()
	}

	s.storeEngine.LoadConfig(config)

	s.servers, s.users, s.allServers = s.storeEngine.Init()

	var l = len(s.allServers)
	if l < _MIN_LEN_SERVER_CHAN {
		l = _MIN_LEN_SERVER_CHAN
	}
	s.AddServerChan = make(chan string, l)
	for server := range s.allServers {
		s.AddServerChan <- server
	}

	s.KickServerChan = make(chan string, l)

	return s
}

func (s *Store) Close() {
	if s.isClosed {
		return
	}
	for atomic.LoadInt64(s.closeCounter) > 0 {
		time.Sleep(10 * time.Millisecond)
	}
	s.isClosed = true
}

func (s *Store) do(f func()) {
	if s.isClosed {
		return
	}
	s.acquire()
	defer s.release()
	f()
}

func (s *Store) acquire() { *s.closeCounter = atomic.AddInt64(s.closeCounter, 1) }

func (s *Store) release() { *s.closeCounter = atomic.AddInt64(s.closeCounter, -1) }

func (s *Store) withWriteLock(f func()) {
	s.rwl.Lock()
	defer s.rwl.Unlock()
	f()
}

func (s *Store) withReadLock(f func()) {
	s.rwl.RLock()
	defer s.rwl.RUnlock()
	f()
}

// user operations
func (s *Store) GetUser(username string) (u *User) {
	s.withReadLock(func() { u = s.users[username] })
	return
}

func (s *Store) UpdatePassword(username string, oldpassword, newpassword string) (err error) {
	s.do(func() {
		s.withReadLock(func() {
			if u, ok := s.users[username]; ok {
				if u.Password != oldpassword {
					err = _ERROR_INCORRECT_PASSWORD
					return
				}
				u.Password = newpassword
				err = s.storeEngine.WriteUser(username, u)
			}
		})
	})
	return
}

func (s *Store) AddUser(username string, password string) (err error) {
	s.do(func() {
		s.withWriteLock(func() {
			if _, ok := s.users[username]; ok {
				err = fmt.Errorf("User %v already exist", username)
			} else {
				s.users[username] = newUser()
				s.users[username].Password = password
				err = s.storeEngine.WriteUser(username, s.users[username])
			}
		})
	})
	return
}

// monitor operations
func (s *Store) DeleteMonitorServer(username string, server string) (err error) {
	s.do(func() {
		s.withWriteLock(func() {
			if u, ok := s.users[username]; !ok {
				err = fmt.Errorf("user %v not exist", username)
			} else {
				if u.MonitorServers[server] {
					delete(u.MonitorServers, server)
					s.allServers[server]--
				}
				if s.allServers[server] <= 0 {
					delete(s.allServers, server)
					s.KickServerChan <- server
				}
				err = s.storeEngine.WriteUser(username, u)
			}
		})
	})
	return
}

func (s *Store) AddMonitorServer(username string, server string) (err error) {
	s.do(func() {
		s.withReadLock(func() {
			if u, ok := s.users[username]; !ok {
				err = fmt.Errorf("User %v not exist", username)
			} else {
				if u.MonitorServers[server] {
					err = fmt.Errorf("%v is already in monitoring list", server)
					return
				}
				u.MonitorServers[server] = true
				if _, ok := s.allServers[server]; !ok {
					s.AddServerChan <- server
				}
				s.allServers[server]++
				err = s.storeEngine.WriteUser(username, u)
			}
		})
	})
	return
}

func (s *Store) AppendPingRet(server string, location string, pr PingRet) (err error) {
	s.do(func() {
		s.withWriteLock(func() {
			if s.allServers[server] <= 0 {
				err = fmt.Errorf("server %v is not exist", server)
				return
			}
			if _, ok := s.servers[ServerAddr(server)]; !ok {
				s.servers[ServerAddr(server)] = make(map[Location][]PingRet)
			}
			if _, ok := s.servers[ServerAddr(server)][Location(location)]; !ok {
				s.servers[ServerAddr(server)][Location(location)] = make([]PingRet, 0)
				// find the longest location, iterate ping results, push into the new location
				var max = -1
				var tLoc Location
				for loc, prs := range s.servers[ServerAddr(server)] {
					if len(prs) > max {
						max = len(prs)
						tLoc = loc
					}
				}
				for _, v := range s.servers[ServerAddr(server)][tLoc] {
					if v.Time == pr.Time {
						break
					}
					s.servers[ServerAddr(server)][Location(location)] =
						append(s.servers[ServerAddr(server)][Location(location)],
							PingRet{Time: v.Time, Ping: "0.000"})
				}
				s.servers[ServerAddr(server)][Location(location)] =
					append(s.servers[ServerAddr(server)][Location(location)], pr)
				err = s.storeEngine.BatchWritePingRets(server, location,
					s.servers[ServerAddr(server)][Location(location)])
			} else {
				s.servers[ServerAddr(server)][Location(location)] =
					append(s.servers[ServerAddr(server)][Location(location)], pr)
				err = s.storeEngine.AppendPingRet(server, location, pr)
			}
		})
	})
	return
}

func (s *Store) GetMonitorResult(username string, server string) (ret map[Location][]PingRet, err error) {
	s.withReadLock(func() {
		if u, ok := s.users[username]; !ok {
			err = fmt.Errorf("User %v not exist", username)
		} else {
			if _, ok := u.MonitorServers[server]; ok {
				ret = s.servers[ServerAddr(server)]
			} else {
				err = fmt.Errorf("You are not monitoring %v", server)
			}
		}
	})
	return
}
