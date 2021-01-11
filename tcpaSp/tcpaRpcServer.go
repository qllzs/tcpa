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

//Tcparmp roc
type Tcparmp struct {
}

//TcpaRequest req
type TcpaRequest struct {
	UeIP   string `json:"ue_ip"`
	OvsIP  string `json:"ovs_ip"`
	TcpaIP string `json:"tcpa_ip"`
}

//Reply reply
type Reply struct {
	ServerIP string `json:"server_ip"`
	Msg      string `json:"msg"`
}

//tcpRPCServer  new server
func tcpRPCServer() {

	rpc.Register(new(Tcparmp))

	addr := ":50052"

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Errorln("tcpRPCServer listen failed:", err)
		return
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Errorln("tcpRPCServer accept failed:", err)
			continue
		}

		go jsonrpc.ServeConn(conn)

	}
}

//CreateGRETunnel ct
func (rpc *Tcparmp) CreateGRETunnel(parms TcpaRequest, reply *string) error {

	var err error
	var out bytes.Buffer
	var stderr bytes.Buffer

	//校验配对IP与本地IP是否一致
	localIP := externalIP("ens160").String()
	if parms.TcpaIP != localIP {
		*reply = "ta IP CRC error"
		return nil
	}

	exec.Command("ip", "link", "delete", "user").Output()
	exec.Command("ip", "link", "delete", "sat").Output()

	//tcparmp-------gre-----------ovs
	userCmd := exec.Command("ip", "link", "add", "user", "type", "gretap", "remote", parms.OvsIP, "local", localIP, "ttl", "255")
	userCmd.Stdout = &out
	userCmd.Stderr = &stderr
	err = userCmd.Run()
	if err != nil {
		*reply = fmt.Sprintf("user:" + fmt.Sprint(err) + ": " + stderr.String())
		fmt.Println("user:" + fmt.Sprint(err) + ": " + stderr.String())
		return nil
	}
	fmt.Println("user:" + "Result: " + out.String())

	//xgw------gre-------------tcparp
	satCmd := exec.Command("ip", "link", "add", "sat", "type", "gretap", "remote", parms.UeIP, "local", localIP, "ttl", "255")
	satCmd.Stdout = &out
	satCmd.Stderr = &stderr
	err = satCmd.Run()
	if err != nil {
		*reply = fmt.Sprintf("sat:" + fmt.Sprint(err) + ": " + stderr.String())
		fmt.Println("sat:" + fmt.Sprint(err) + ": " + stderr.String())
		return nil
	}
	fmt.Println("sat:" + "Result: " + out.String())

	//启动tcp加速器
	statrCmd := exec.Command("/opt/nkt/tcpa/start.sh")
	statrCmd.Stdout = &out
	statrCmd.Stderr = &stderr
	statrCmd.Run()

	*reply = "succeed"

	return nil
}

//ReleaseGRETunnel rt
func (rpc *Tcparmp) ReleaseGRETunnel(tcpaIP string, reply *string) error {

	var err error
	var out bytes.Buffer
	var stderr bytes.Buffer

	//校验配对IP与本地IP是否一致
	localIP := externalIP("ens160").String()
	if tcpaIP != localIP {
		*reply = "ta IP CRC error"
		return nil
	}

	//tcparmp-------gre-----------ovs
	delUserCmd := exec.Command("ip", "link", "delete", "user")
	delUserCmd.Stdout = &out
	delUserCmd.Stderr = &stderr
	err = delUserCmd.Run()
	if err != nil {
		*reply = fmt.Sprintf("del user:" + fmt.Sprint(err) + ": " + stderr.String())
		fmt.Println("del user:" + fmt.Sprint(err) + ": " + stderr.String())
		return nil
	}
	fmt.Println("del user:" + "Result: " + out.String())

	//xgw------gre-------------tcparp
	delSatCmd := exec.Command("ip", "link", "delete", "sat")
	delSatCmd.Stdout = &out
	delSatCmd.Stderr = &stderr
	err = delSatCmd.Run()
	if err != nil {
		*reply = fmt.Sprintf("del sat:" + fmt.Sprint(err) + ": " + stderr.String())
		fmt.Println("del sat:" + fmt.Sprint(err) + ": " + stderr.String())
		return nil
	}
	fmt.Println("del sat:" + "Result: " + out.String())

	//停止tcp加速器
	stopCmd := exec.Command("/opt/nkt/tcpa/stop.sh")
	stopCmd.Stdout = &out
	stopCmd.Stderr = &stderr
	stopCmd.Run()

	*reply = "succeed"

	return nil
}
