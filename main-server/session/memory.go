// memory store
package session

import (
	"sync"
	"time"
)

const STORE_MEMORY = "memory"

type memory struct {
	data       map[interface{}]interface{}
	lastUpdate time.Time

	sid string
	rwl sync.RWMutex
}

func init() {
	Register(STORE_MEMORY, newMemoryStore)
}

func newMemoryStore(sid string, conf string) SessionStore {
	return &memory{
		data:       make(map[interface{}]interface{}),
		lastUpdate: time.Now(),
		sid:        sid,
	}
}

func (m *memory) withReadLock(f func()) {
	m.rwl.RLock()
	defer m.rwl.RUnlock()
	f()
}

func (m *memory) withWriteLock(f func()) {
	m.rwl.Lock()
	defer m.rwl.Unlock()
	f()
}

func (m *memory) SId() string { return m.sid }

func (m *memory) Expire() error { return nil }

func (m *memory) Update() error {
	m.lastUpdate = time.Now()
	return nil
}

func (m *memory) Init() map[string]SessionStore { return make(map[string]SessionStore) }

func (m *memory) Iterate(f func(key, val interface{})) {
	m.withReadLock(func() {
		for key, val := range m.data {
			f(key, val)
		}
	})
}

func (m *memory) Set(key, val interface{}) error {
	m.withWriteLock(func() {
		m.data[key] = val
		m.update()
	})
	return nil
}

func (m *memory) Get(key interface{}) (val interface{}) {
	m.withReadLock(func() { val = m.data[key] })
	return
}

func (m *memory) Delete(key interface{}) error {
	m.withWriteLock(func() {
		delete(m.data, key)
		m.update()
	})
	return nil
}

func (m *memory) update() { m.lastUpdate = time.Now() }

func (m memory) LastUpdate() time.Time { return m.lastUpdate }
