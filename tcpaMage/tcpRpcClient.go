package main

import (
	"net"
	"net/rpc/jsonrpc"
	"strings"

	log "github.com/sirupsen/logrus"
)

func tcpaRPCClient(taIP string) error {

	ips := strings.Split(taIP, ":")
	taServerIP := ips[0]

	conn, err := net.Dial("tcp", taServerIP+":50052")
	if err != nil {
		gLoger.WithFields(log.Fields{"taServerIP": taServerIP}).Errorln("dail ta  rpc Server failed")
		return err
	}

	ta := tcpamObj.tcpaMap[taIP]

	ta.cli = jsonrpc.NewClient(conn)
	if ta.cli == nil {
		gLoger.WithFields(log.Fields{"taServerIP": taServerIP}).Errorln("create client for ta rpc Server failed")
		return err
	}

	gLoger.WithFields(log.Fields{"taServerIP": taServerIP}).Infoln("create client for ta rpc Server")

	return nil
}
