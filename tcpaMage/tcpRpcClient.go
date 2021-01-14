package main

import (
	"net"
	"net/rpc/jsonrpc"

	log "github.com/sirupsen/logrus"
)

func tcpaRPCClient(taServerIP string) error {

	conn, err := net.Dial("tcp", taServerIP+":50052")
	if err != nil {
		gLoger.WithFields(log.Fields{"taServerIP": taServerIP}).Errorln("dail ta  rpc Server failed")
		return err
	}

	ta := tcpamObj.tcpaMap[taServerIP]

	ta.cli = jsonrpc.NewClient(conn)
	if ta.cli == nil {
		gLoger.WithFields(log.Fields{"taServerIP": taServerIP}).Errorln("create client for ta rpc Server failed")
		return err
	}

	ta.isIdle = true //tcpa标识空闲可用

	gLoger.WithFields(log.Fields{"tcpa ip": ta.tcpaIP, "tcpa isIdle": ta.isIdle}).Infoln("create client for ta rpc Server")

	return nil
}
