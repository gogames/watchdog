// file store
// with memory cache :-)
package session

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hprose/hprose-go/hprose"
)

const STORE_FILE = "file"

type file struct {
	Data            map[interface{}]interface{}
	LastUpdatedTime time.Time

	sid string
	rwl sync.RWMutex

	path string
}

func init() {
	Register(STORE_FILE, newFileStore)
}

func newFileStore(sid string, conf string) SessionStore {
	f := &file{
		Data:            make(map[interface{}]interface{}),
		LastUpdatedTime: time.Now(),
		sid:             sid,
	}
	f.parseConf(conf)
	return f
}

func (f *file) marshal() []byte {
	b, _ := hprose.Marshal(f)
	return b
}

func (f file) getFileName() string { return fmt.Sprintf("%v/%v", f.path, f.sid) }

func (f *file) withReadLock(fc func()) {
	f.rwl.RLock()
	defer f.rwl.RUnlock()
	fc()
}

func (f *file) withWriteLock(fc func()) {
	f.rwl.Lock()
	defer f.rwl.Unlock()
	fc()
}

func (f *file) SId() string { return f.sid }

func (f *file) parseConf(conf string) {
	m := make(map[string]string)
	if err := json.Unmarshal([]byte(conf), &m); err != nil {
		panic(err)
	}
	var ok bool
	f.path, ok = m["path"]
	if !ok {
		panic("has no path specified in conf")
	}
}

func (f *file) Init() map[string]SessionStore {
	if _, err := os.Stat(f.path); os.IsNotExist(err) {
		if err = os.Mkdir(f.path, os.ModePerm); err != nil {
			panic(fmt.Errorf("can not mkdir: %v", err))
		}
	}

	ret := make(map[string]SessionStore)
	if err := filepath.Walk(f.path, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		var ff file
		if err := hprose.Unmarshal(b, &ff); err != nil {
			return err
		}
		ff.sid = info.Name()
		ff.path = f.path
		ret[info.Name()] = &ff
		return nil
	}); err != nil {
		panic(fmt.Sprintf("can not init from file: %v\n", err))
	}
	return ret
}

func (f *file) Iterate(fc func(key, val interface{})) {
	f.withReadLock(func() {
		for key, val := range f.Data {
			fc(key, val)
		}
	})
}

func (f *file) Set(key, val interface{}) (err error) {
	f.withWriteLock(func() {
		f.Data[key] = val
		f.update()
		err = ioutil.WriteFile(f.getFileName(), f.marshal(), os.ModePerm)
	})
	return
}

func (f *file) Get(key interface{}) (val interface{}) {
	f.withReadLock(func() { val = f.Data[key] })
	return
}

func (f *file) Delete(key interface{}) (err error) {
	f.withWriteLock(func() {
		delete(f.Data, key)
		f.update()
		err = ioutil.WriteFile(f.getFileName(), f.marshal(), os.ModePerm)
	})
	return
}

func (f *file) Expire() error { return os.Remove(f.getFileName()) }

func (f *file) Update() (err error) {
	f.withReadLock(func() {
		f.update()
		err = ioutil.WriteFile(f.getFileName(), f.marshal(), os.ModePerm)
	})
	return os.Remove(f.getFileName())
}

func (f *file) update() { f.LastUpdatedTime = time.Now() }

func (f file) LastUpdate() time.Time { return f.LastUpdatedTime }
