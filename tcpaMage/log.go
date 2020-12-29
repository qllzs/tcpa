package main

import (
	"os"

	log "github.com/sirupsen/logrus"
)

var gLoger *log.Entry

func init() {

	gLog := log.New()

	//create log file
	logfile, err := os.Create("/opt/nkt/tcpaMage/log/tcparm.log")
	if err != nil {
		gLog.WithFields(log.Fields{"err": err}).Errorln("/opt/nkt/tcpaMage/log/tcparm.log")
		return
	}

	gLog.SetFormatter(&log.JSONFormatter{})

	gLog.SetOutput(logfile)
	//gLog.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	gLog.SetLevel(log.InfoLevel)

	gLoger = gLog.WithFields(log.Fields{"own": "main"})

}
