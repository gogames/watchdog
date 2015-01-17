package main

import (
	"fmt"
	"time"

	"github.com/gogames/watchdog/main-server/session"
)

var sess *session.Session

func initSession() {
	sess = session.NewSession(time.Hour*2, fmt.Sprintf(`{"path":"%s"}`, *flagSessionDirectory)).SetProvider(session.STORE_FILE)
}
