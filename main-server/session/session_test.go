package session

import (
	"testing"
	"time"
)

func Test_Session(t *testing.T) {
	s := NewSession(time.Second, `{"path":"dir"}`).SetProvider(STORE_FILE)

	var err error
	sid, key, val := "", "helloKey", "helloWorld"

	sid, err = s.Set(sid, key, val)
	if err != nil {
		t.Error(err)
	}
	t.Log(sid)

	if s.Get(sid, key).(string) != val {
		t.Errorf("the value is not %v", val)
	}

	time.Sleep(2 * time.Second)

	if s.Get(sid, key) != nil {
		t.Error("the value should be nil interface{}")
	}

	nsid, err := s.Set("", "ha", "lo")
	if err != nil {
		t.Error(err)
	}
	if s.Get(nsid, "ha").(string) != "lo" {
		t.Error("should be lo")
	}

	s.Close()
	// s.Close()
	// s.Close()

	t.Log(s.Set(nsid, "what", "the"))

	// t.Log(s.Get(nsid, "what"))
}
