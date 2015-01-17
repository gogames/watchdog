package store

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	_NEW_LINE   = "\n"
	ENGINE_FILE = "file"
)

type fileEngine struct {
	serversDir, usersDir string
	cursor               string

	servers    Servers
	users      Users
	allServers map[string]int64
}

func init() {
	Register(ENGINE_FILE, newFileEngine)
}

func newFileEngine() StoreEngine {
	return &fileEngine{
		servers:    make(Servers),
		users:      make(Users),
		allServers: make(map[string]int64),
	}
}

func (f *fileEngine) LoadConfig(s string) {
	m := make(map[string]string)
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		panic(err)
	}
	var ok bool
	f.serversDir, ok = m["serversDir"]
	if !ok {
		panic("should config serversDir")
	}
	f.usersDir, ok = m["usersDir"]
	if !ok {
		panic("should config usersDir")
	}
}

func (f *fileEngine) WriteUser(username string, u *User) error {
	return ioutil.WriteFile(f.getUserFilePath(username), u.marshal(), os.ModePerm)
}

func (f *fileEngine) AppendPingRet(server string, location string, pr PingRet) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	f.notExistThenMkdir(f.getServerDir(server))
	return f.appendFile(f.getServerFilePath(server, location), pr.marshal(), os.ModePerm)
}

func (f *fileEngine) BatchWritePingRets(server string, location string, prs []PingRet) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	f.notExistThenMkdir(f.getServerDir(server))
	bs := bytes.NewBuffer(make([]byte, 0))
	for _, pr := range prs {
		_, err = bs.Write(pr.marshal())
		if err != nil {
			return
		}
	}
	return f.appendFile(f.getServerFilePath(server, location), bs.Bytes(), os.ModePerm)
}

func (f *fileEngine) Init() (Servers, Users, map[string]int64) {
	defer func() {
		f.servers = nil
		f.users = nil
		f.allServers = nil
	}()

	f.notExistThenMkdir(f.serversDir)
	f.notExistThenMkdir(f.usersDir)

	f.cursor = f.serversDir
	if err := filepath.Walk(f.serversDir, f.serversWalkerFunc); err != nil {
		panic("can not walk servers")
	}

	f.cursor = f.usersDir
	if err := filepath.Walk(f.usersDir, f.usersWalkerFunc); err != nil {
		panic("can not walk users")
	}

	return f.servers, f.users, f.allServers
}

func (f *fileEngine) getUserFilePath(username string) string {
	return fmt.Sprintf("%v/%v", f.usersDir, username)
}

func (f *fileEngine) getServerFilePath(serverAddr, location string) string {
	return fmt.Sprintf("%v/%v/%v", f.serversDir, serverAddr, location)
}

func (f *fileEngine) getServerDir(serverAddr string) string {
	return fmt.Sprintf("%v/%v", f.serversDir, serverAddr)
}

func (f *fileEngine) serversWalkerFunc(path string, file os.FileInfo, err error) error {
	if file.Name() == f.cursor {
		return nil
	}

	if file.IsDir() {
		f.cursor = file.Name()
		if f.servers[ServerAddr(f.cursor)] == nil {
			f.servers[ServerAddr(f.cursor)] = make(map[Location][]PingRet)
		}
		return filepath.Walk(path, f.serversWalkerFunc)
	} else {
		f.servers[ServerAddr(f.cursor)][Location(file.Name())] = f.getPingRetsFromPath(path)
	}
	return nil
}

func (f *fileEngine) getPingRetsFromPath(path string) []PingRet {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	ps := make([]PingRet, 0)
	for _, v := range bytes.Split(bs, []byte(_NEW_LINE)) {
		var p PingRet
		if len(v) == 0 {
			continue
		}
		if err := json.Unmarshal(v, &p); err != nil {
			panic(err)
		}
		ps = append(ps, p)
	}
	return ps
}

func (f *fileEngine) usersWalkerFunc(path string, file os.FileInfo, err error) error {
	if file.Name() == f.cursor {
		return nil
	}

	if !file.IsDir() {
		u := f.getUserFromPath(path)
		f.users[file.Name()] = u
		for server := range u.MonitorServers {
			f.allServers[server]++
		}
	}
	return nil
}

func (f *fileEngine) getUserFromPath(path string) *User {
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	u := newUser()
	if err := json.Unmarshal(bs, u); err != nil {
		panic(err)
	}
	return u
}

func (f *fileEngine) notExistThenMkdir(dir string) error {
	if exist, err := f.isDirExist(dir); err != nil {
		return err
	} else if !exist {
		if err = f.mkdir(dir); err != nil {
			return err
		}
	}
	return nil
}

func (f *fileEngine) isDirExist(dir string) (exist bool, err error) {
	file, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
	} else {
		if !file.IsDir() {
			err = fmt.Errorf("%v is not directory", file.Name())
		}
		exist = true
	}
	return
}

func (f *fileEngine) mkdir(dir string) error { return os.Mkdir(dir, os.ModePerm) }

func (f *fileEngine) isFileExist(dir string) (exist bool, err error) {
	file, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
	} else {
		if file.IsDir() {
			err = fmt.Errorf("%s is not file", file.Name())
		}
		exist = true
	}
	return
}

func (f *fileEngine) appendFile(path string, data []byte, perm os.FileMode) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	n, err := file.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := file.Close(); err == nil {
		err = err1
	}
	return err
}
