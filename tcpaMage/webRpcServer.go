package main

import (
	"net"
	"net/rpc"

	log "github.com/sirupsen/logrus"
)

//Web  rpc struct
type Web struct {
}

func init() {

	rpc.Register(new(Web))

	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		gLoger.WithFields(log.Fields{"err": err.Error()}).Errorln("init web rpc sever listen failed")
		return
	}
	gLoger.WithFields(log.Fields{"ip": lis.Addr().String()}).Errorln("init web rpc sever listen at")

}
