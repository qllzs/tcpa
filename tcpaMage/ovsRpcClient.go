package main

import (
	"errors"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"

	log "github.com/sirupsen/logrus"
)

var ovsRPCCli *rpc.Client

func init() {
	ovsMap := GViperCfg.GetStringMapString("ovs")

	for ovsRPCIP, ovsIP := range ovsMap {
		cli, err := ovsRPCClient(ovsRPCIP)
		if err != nil {
			gLoger.WithFields(log.Fields{"ovs rpc ip": ovsRPCIP, "ovs ip": ovsIP}).Errorln(err.Error())
		}

		tcpamObj.ovsNum++
		var ov ovs
		ov.isIdle = false
		ov.ovsCli = cli
		ov.ueNum = 0
		ov.ovsIP = ovsIP
		ov.rpcIP = ovsRPCIP
		tcpamObj.ovsMap[ovsIP] = &ov

		gLoger.WithFields(log.Fields{"ovs rpc ip": ovsRPCIP, "ovs ip": ovsIP}).Infoln("create client for ovs rpc Server succeed")
	}

}

func ovsRPCClient(ovsRPCIP string) (*rpc.Client, error) {

	var err error
	var conn net.Conn

	conn, err = net.Dial("tcp", ovsRPCIP+":50054")
	if err != nil {
		return nil, err
	}

	cli := jsonrpc.NewClient(conn)
	if ovsRPCCli == nil {
		return nil, errors.New("new ovs rpc client failed")
	}

	return cli, nil
}
