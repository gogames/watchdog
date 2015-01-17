package main

import (
	"flag"

	"github.com/astaxie/beego/logs"
)

var (
	flagLogFilePath        = flag.String("log", "/var/log/servermonitor/mainserver/logfile.log", "log file")
	flagLogLevel           = flag.Int("level", logs.LevelDebug, "log level")
	flagPingNodeServerPort = flag.Int("pingport", 8563, "port to invoke ping node")
	flagManagerPort        = flag.Int("managerport", 8773, "port to run manager")
	flagMainServerPort     = flag.Int("port", 8683, "port to run main server")
	flagPingInterval       = flag.Int("pinginterval", 60, "number of seconds to kick a ping node")
	flagServersPath        = flag.String("serverspath", "storeServers", "path to store ping results of servers")
	flagUsersPath          = flag.String("userspath", "storeUsers", "path to store user information")
	flagPingFrequence      = flag.Int("pingfreq", 10, "monitor the server by ping every ping frequence minutes")
	flagSessionDirectory   = flag.String("sessiondir", "sessionDirectory", "path to store the sessions")
)

func initFlag() { flag.Parse() }
