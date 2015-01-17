// parse flags
package main

import (
	"flag"

	"github.com/astaxie/beego/logs"
)

var (
	flagMainServerAddress = flag.String("addr", "127.0.0.1:8773", "network address of main server")
	flagLocation          = flag.String("location", "local", "location of the ping node")
	flagLogFilePath       = flag.String("log", "/var/log/servermonitor/pingnode/logfile.log", "location of the ping node")
	flagLogLevel          = flag.Int("level", logs.LevelDebug, "log level according to RFC5424, default debug level")
)

func initFlag() {
	flag.Parse()
}
