package main

import (
	"fmt"

	"github.com/astaxie/beego/logs"
)

var logger = logs.NewLogger(1 << 15)

func initLogger() {
	if err := logger.SetLogger("file", fmt.Sprintf(`{"filename":"%s"}`, *flagLogFilePath)); err != nil {
		panic("can not set logger: " + err.Error())
	}
	logger.SetLevel(*flagLogLevel)
}
