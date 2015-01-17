package pingClientManager

import (
	"testing"
	"time"
)

func Test_PCM(t *testing.T) {
	pcm := NewPingClientManager(3)

	if err := pcm.Register("Hong Kong", "http://whatever.com:9090/"); err != nil {
		t.Error(err)
	}

	pcm.Iterate(func(location string, _ PingClient) {
		if location != "Hong Kong" {
			t.Error("Register failed")
		}
	})

	time.Sleep(2 * time.Second)
	// pcm.Ping("Hong Kong")
	t.Log("after 2 seconds")

	pcm.Iterate(func(location string, _ PingClient) {
		if location == "Hong Kong" {
			t.Error("kick failed")
		}
	})
}
