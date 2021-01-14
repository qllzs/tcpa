package main

import (
	"bytes"
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

//TcpaOvs ovs
type TcpaOvs struct {
}

//Reply reply
type Reply struct {
	ServerIP string `json:"server_ip"`
	Msg      string `json:"msg"`
}

func ovsRPCServer() {

	rpc.Register(new(TcpaOvs))

	lis, err := net.Listen("tcp", ":50054")
	if err != nil {
		log.Errorln("ovs rpc server listen failed:", err.Error())
		return
	}
	log.WithFields(log.Fields{"ovs rpc server listen ip:": lis.Addr().String()}).Infoln("ovs rpc server listen at:")

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Errorln("ovs rpc server accept failed")
			continue
		}

		log.WithFields(log.Fields{"ovs rpc server, tcpa manage ip": conn.RemoteAddr().String()}).Infoln("ovs rpc  server accept")

		go jsonrpc.ServeConn(conn)
	}

}

//CreateOvsGRETunnel   create
func (rpc *TcpaOvs) CreateOvsGRETunnel(tcpaIP string, reply *string) error {

	var err error
	var out bytes.Buffer
	var stderr bytes.Buffer

	//ovs-vsctl add-port tcpa_ovs_br 30.254.253.1 -- set interface 30.254.253.1 type=gre options:remote_ip=30.254.253.1
	cmd := exec.Command("ovs-vsctl", "add-port", "tcpa_ovs_br", tcpaIP, "--", "set", "interface", tcpaIP, "type=gre", "options:remote_ip="+tcpaIP)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		*reply = fmt.Sprintf("ovs crt:" + fmt.Sprint(err) + ": " + stderr.String())
		log.Errorln("ovs crt:" + fmt.Sprint(err) + ": " + stderr.String())
		return nil
	}
	log.Infoln("ovs crt:" + "Result: " + out.String())

	*reply = "succeed"
	return nil
}

//ReleaseOvsGRETunnel release
func (rpc *TcpaOvs) ReleaseOvsGRETunnel(tcpaIP string, reply *string) error {

	var err error
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("ovs-vsctl", "del-port", tcpaIP)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		*reply = fmt.Sprintf("ovs del:" + fmt.Sprint(err) + ": " + stderr.String())
		log.Errorln("ovs del:" + fmt.Sprint(err) + ": " + stderr.String())
		return nil
	}
	log.Infoln("ovs del:" + "Result: " + out.String())

	*reply = "succeed"
	return nil
}
