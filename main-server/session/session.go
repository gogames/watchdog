package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const (
	_SID_LENGTH = 1 << 4
)

type SessionStore interface {
	Iterate(func(key, val interface{}))
	Set(key, val interface{}) error
	Get(interface{}) interface{}
	Update() error
	Delete(interface{}) error
	LastUpdate() time.Time
	SId() string
	Expire() error
	Init() map[string]SessionStore
}

var sessions = make(map[string]func(string, string) SessionStore)

func Register(sessionStoreName string, f func(string, string) SessionStore) {
	if _, ok := sessions[sessionStoreName]; ok {
		panic(fmt.Sprintf("can not register %v twice", sessionStoreName))
	}
	sessions[sessionStoreName] = f
}

type Session struct {
	sessions map[string]SessionStore

	storeProvider func(string, string) SessionStore
	gcDuration    time.Duration

	closeCounter *int64
	closeChan    chan struct{}
	isClosed     bool

	conf string

	rwl sync.RWMutex
}

func NewSession(gcDuration time.Duration, conf string) *Session {
	s := &Session{
		gcDuration:   gcDuration,
		closeCounter: new(int64),
		closeChan:    make(chan struct{}),
		conf:         conf,
	}
	go s.gc()
	return s
}

func (s *Session) Close() {
	if s.isClosed {
		return
	}
	for atomic.LoadInt64(s.closeCounter) > 0 {
		time.Sleep(10 * time.Millisecond)
	}
	s.isClosed = true
	s.closeChan <- struct{}{}
	s.sessions = nil
	return
}

func (s *Session) acquire() { atomic.AddInt64(s.closeCounter, 1) }

func (s *Session) release() { atomic.AddInt64(s.closeCounter, -1) }

func (s *Session) do(f func()) {
	if s.isClosed {
		return
	}
	s.acquire()
	defer s.release()
	f()
}

func (s *Session) gc() {
	for {
		select {
		case t := <-time.Tick(time.Minute):
			s.do(func() {
				s.withWriteLock(func() {
					for sid, ss := range s.sessions {
						if t.Sub(ss.LastUpdate()) > s.gcDuration {
							ss.Expire()
							delete(s.sessions, sid)
						}
					}
				})
			})
		case <-s.closeChan:
			return
		}
	}
}

func (s *Session) SetProvider(sessionStoreName string) *Session {
	if _, ok := sessions[sessionStoreName]; !ok {
		panic(fmt.Sprintf("session store %v is not registered", sessionStoreName))
	}
	s.storeProvider = sessions[sessionStoreName]
	s.sessions = s.storeProvider("", s.conf).Init()
	return s
}

func (s *Session) Update(sid string) (err error) {
	s.do(func() {
		s.withReadLock(func() {
			sp, ok := s.sessions[sid]
			if ok {
				err = sp.Update()
			}
		})
	})
	return
}

func (s *Session) Set(sid string, key, val interface{}) (sessionId string, err error) {
	s.do(func() {
		var cont bool
		s.withReadLock(func() {
			sp, ok := s.sessions[sid]
			if ok {
				cont = ok
				err = sp.Set(key, val)
			}
		})
		if !cont {
			s.withWriteLock(func() {
				sid = s.newSId()
				s.sessions[sid] = s.storeProvider(sid, s.conf)
				err = s.sessions[sid].Set(key, val)
			})
		}
		sessionId = sid
	})
	return
}

func (s *Session) Get(sid string, key interface{}) (val interface{}) {
	s.withReadLock(func() {
		if sp, ok := s.sessions[sid]; ok {
			val = sp.Get(key)
		}
	})
	return
}

func (s *Session) Delete(sid string, key interface{}) (err error) {
	s.do(func() {
		s.withReadLock(func() {
			if sp, ok := s.sessions[sid]; ok {
				err = sp.Delete(key)
			}
		})
	})
	return
}

func (s *Session) Iterate(f func(sid string, ss SessionStore)) {
	s.do(func() {
		s.withReadLock(func() {
			for sid, ss := range s.sessions {
				f(sid, ss)
			}
		})
	})
}

func (s *Session) Expire(sid string) (err error) {
	s.do(func() {
		s.withWriteLock(func() {
			if ss, ok := s.sessions[sid]; ok {
				err = ss.Expire()
			}
			delete(s.sessions, sid)
		})
	})
	return
}

func (s *Session) withReadLock(f func()) {
	s.rwl.RLock()
	defer s.rwl.RUnlock()
	f()
}

func (s *Session) withWriteLock(f func()) {
	s.rwl.Lock()
	defer s.rwl.Unlock()
	f()
}

func (s *Session) newSId() string {
	for {
		if sid := s.randSId(); sid != "" {
			if _, ok := s.sessions[sid]; ok {
				continue
			}
			return sid
		}
	}
}

func (s *Session) randSId() string {
	for {
		b := make([]byte, _SID_LENGTH)
		n, err := rand.Read(b)
		if err != nil && n != _SID_LENGTH {
			log.Printf("can not rand.Read %v", err)
			return ""
		}
		return hex.EncodeToString(b)
	}
}
