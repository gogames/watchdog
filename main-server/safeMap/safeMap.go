package safeMap

import "sync"

type safeMap struct {
	data map[interface{}]interface{}
	rwl  sync.RWMutex
}

func NewSafeMap() *safeMap { return &safeMap{data: make(map[interface{}]interface{})} }

// set if key does not exist
func (s *safeMap) Set(key, val interface{}) (success bool) {
	s.rwl.Lock()
	defer s.rwl.Unlock()
	if s.data[key] == nil {
		s.data[key] = val
		success = true
	}
	return
}

// update if key exist
func (s *safeMap) Update(key, val interface{}) (success bool) {
	s.rwl.Lock()
	defer s.rwl.Unlock()
	if s.data[key] != nil {
		s.data[key] = val
		success = true
	}
	return
}

// set key -> val no matter the key exist or not
func (s *safeMap) SetOrUpdate(key, val interface{}) (exist bool) {
	s.rwl.Lock()
	defer s.rwl.Unlock()
	exist = s.data[key] != nil
	s.data[key] = val
	return
}

// get the key
func (s *safeMap) Get(key interface{}) interface{} {
	s.rwl.RLock()
	defer s.rwl.RUnlock()
	return s.data[key]
}

func (s *safeMap) Delete(key interface{}) {
	s.rwl.Lock()
	defer s.rwl.Unlock()
	delete(s.data, key)
}
