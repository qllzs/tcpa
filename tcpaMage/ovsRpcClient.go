package main

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"

	log "github.com/sirupsen/logrus"
)

var ovsRPCCli *rpc.Client
var ovsCh chan struct{}

func ovsRPCClient() error {

	var err error
	var conn net.Conn
	ovsCh = make(chan struct{})

	for {

		ovsIP := GViperCfg.GetString("ovs_rpc_ip")
		conn, err = net.Dial("tcp", ovsIP+":50054")
		if err != nil {
			time.Sleep(time.Second)
			gLoger.WithFields(log.Fields{"ovsIP": ovsIP}).Errorln("connect ovs failed")
			continue
		}

		ovsRPCCli = jsonrpc.NewClient(conn)
		if ovsRPCCli == nil {
			gLoger.Errorln("new ovs rpc client failed")
			continue
		}
		gLoger.WithFields(log.Fields{"ovs IP": conn.RemoteAddr().String()}).Infoln("creeat client for ovs rpc Server")
		<-ovsCh
	}

	//return nil
}
