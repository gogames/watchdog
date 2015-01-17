package store

import (
	"testing"
	"time"
)

// test
func testStore(t *testing.T) {
	s := NewStore().SetStoreEngine(ENGINE_FILE, `{"serversDir":"tmpServers","usersDir":"tmpUsers"}`)

	// if err := s.AddUser("newuser", "HELLO"); err != nil {
	// 	t.Error(err)
	// }

	if err := s.AddMonitorServer("newuser", "baidu.com"); err != nil {
		t.Error(err)
	}

	// if err := s.AddMonitorServer("newuser", "google.com"); err != nil {
	// 	t.Error(err)
	// }

	if err := s.AddMonitorServer("newuser", "yahoo.com"); err != nil {
		t.Error(err)
	}

	if err := s.DeleteMonitorServer("newuser", "baidu.com"); err != nil {
		t.Error(err)
	}

	if err := s.DeleteMonitorServer("newuser", "yahoo.com"); err != nil {
		t.Error(err)
	}

	if err := s.AppendPingRet("google.com", "Hong Kong", PingRet{0.392, time.Now().Unix()}); err != nil {
		t.Error(err)
	}

	// t.Log(s.servers)
	// t.Log(s.allServers)
	ret, err := s.GetMonitorResult("newuser", "google.com")
	if err != nil {
		t.Error(err)
	}

	for location, ps := range ret {
		for _, p := range ps {
			t.Logf("%v -> %v", location, p)
		}
	}

	s.Close()
	t.Log(s.close)
}

func Test_Store(t *testing.T) { testStore(t) }
