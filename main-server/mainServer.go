package main

import (
	"fmt"
	"net/http"
	"reflect"
	"syscall"

	"github.com/gogames/watchdog/main-server/store"
	"github.com/gogames/utils/signal"
	"github.com/hprose/hprose-go/hprose"
)

const (
	_SESS_KEY_USERNAME = "username"
)

type (
	mainServerStub struct{}
)

// main server stub

// without signed in
// auto sign the user in
func (mainServerStub) Register(username, password string) (sid, un string, err error) {
	if err = storeEngine.AddUser(username, password); err != nil {
		return
	}
	sid, err = sess.Set("", _SESS_KEY_USERNAME, username)
	un = username
	return
}

func (mainServerStub) Login(username, password string) (sid, un string, err error) {
	un = username
	if u := storeEngine.GetUser(username); u == nil {
		err = fmt.Errorf("user %v does not exist", username)
	} else if u.Password != password {
		err = fmt.Errorf("incorrect password")
	} else {
		sid, err = sess.Set("", _SESS_KEY_USERNAME, username)
	}
	return
}

// update session life
func (mainServerStub) UpdatePassword(sid, username, oldP, newP string) (signedIn bool, err error) {
	if v := sess.Get(sid, _SESS_KEY_USERNAME); v != nil {
		un, ok := v.(string)
		if !ok {
			logger.Debug("the username is not type string, but %v", reflect.TypeOf(v).Name())
			err = fmt.Errorf("Unexpected runtime error")
		} else if un == username {
			if err = storeEngine.UpdatePassword(username, oldP, newP); err != nil {
				return
			}
			err = sess.Update(sid)
			signedIn = true
		}
	}
	return
}

// update session life
func (mainServerStub) AddServer(sid, username, server string) (signedIn bool, err error) {
	if v := sess.Get(sid, _SESS_KEY_USERNAME); v != nil {
		un, ok := v.(string)
		if !ok {
			logger.Debug("the username is not type string, but %v", reflect.TypeOf(v).Name())
			err = fmt.Errorf("Unexpected runtime error")
		} else if un == username {
			if err = storeEngine.AddMonitorServer(username, server); err != nil {
				return
			}
			err = sess.Update(sid)
			signedIn = true
		}
	}
	return
}

// update session life
func (mainServerStub) DelServer(sid, username, server string) (signedIn bool, err error) {
	if v := sess.Get(sid, _SESS_KEY_USERNAME); v != nil {
		un, ok := v.(string)
		if !ok {
			logger.Debug("the username is not type string, but %v", reflect.TypeOf(v).Name())
			err = fmt.Errorf("Unexpected runtime error")
		} else if un == username {
			if err = storeEngine.DeleteMonitorServer(username, server); err != nil {
				return
			}
			err = sess.Update(sid)
			signedIn = true
		}
	}
	return
}

func (mainServerStub) GetMonitorResult(sid, username, server string) (ret map[store.Location][]store.PingRet, signedIn bool, err error) {
	if v := sess.Get(sid, _SESS_KEY_USERNAME); v != nil {
		un, ok := v.(string)
		if !ok {
			logger.Debug("the username is not type string, but %v", reflect.TypeOf(v).Name())
			err = fmt.Errorf("Unexpected runtime error")
		} else if un == username {
			ret, err = storeEngine.GetMonitorResult(username, server)
			signedIn = true
		}
	}
	return
}

func (mainServerStub) GetUser(sid, username string) (u store.User, signedIn bool, err error) {
	if v := sess.Get(sid, _SESS_KEY_USERNAME); v != nil {
		un, ok := v.(string)
		if !ok {
			logger.Debug("the username is not type string, but %v", reflect.TypeOf(v).Name())
			err = fmt.Errorf("Unexpected runtime error")
		} else if un == username {
			up := storeEngine.GetUser(username)
			if up == nil {
				err = fmt.Errorf("User %v does not exist", username)
			} else {
				u = *up
			}
			signedIn = true
		}
	}
	return
}

func (mainServerStub) Logout(sid, username string) (signedIn bool, err error) {
	if v := sess.Get(sid, _SESS_KEY_USERNAME); v != nil {
		un, ok := v.(string)
		if !ok {
			logger.Debug("the username is not type string, but %v", reflect.TypeOf(v).Name())
			err = fmt.Errorf("Unexpected runtime error")
		} else if un == username {
			err = sess.Expire(sid)
			signedIn = true
		}
	}
	return
}

var mainServer = hprose.NewHttpService()

func initMainServer() {
	mainServer.AddMethods(new(mainServerStub))
	mainServer.GetEnabled = true
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%v", *flagMainServerPort), mainServer); err != nil {
			logger.Emergency("can not listen and serve main server: %v", err)
			if err = signal.Signal(syscall.SIGQUIT); err != nil {
				panic(fmt.Errorf("can not signal the current process: %v", err))
			}
		}
	}()
}
