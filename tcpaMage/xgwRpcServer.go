package main

import (
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"

	log "github.com/sirupsen/logrus"
)

//Xgwka  rpc struct
type Xgwka struct {
}

//TcpaRequest req
type TcpaRequest struct {
	UeIP   string `json:"ue_ip"`
	OvsIP  string `json:"ovs_ip"`
	TcpaIP string `json:"tcpa_ip"`
}

//Request req
type Request struct {
	UeIP   string `json:"ue_ip"`
	TcpaIP string `json:"tcpa_ip"`
	XgwIP  string `json:"xgw_ip"`
}

//Reply reply
type Reply struct {
	ServerIP string `json:"server_ip"`
	Msg      string `json:"msg"`
}

func init() {

	rpc.Register(new(Xgwka))

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		gLoger.WithFields(log.Fields{"err": err.Error()}).Errorln("xgwRPCServer xgw rpc sever listen failed")
		return
	}
	gLoger.WithFields(log.Fields{"ip": lis.Addr().String()}).Errorln("xgwRPCServer xgw rpc sever listen at")

	go func() {
		for {
			conn, err := lis.Accept()
			if err != nil {
				gLoger.WithFields(log.Fields{"err": err.Error()}).Errorln("xgw sever accept failed")
				continue
			}

			xgwIP := getXgwIPFromRPCConn(conn)

			tcpamObj.xgwNum++
			tcpamObj.xgwMap[xgwIP] = &[]string{}

			gLoger.WithFields(log.Fields{"addr": xgwIP}).Infoln("xgw sever accept at")

			go jsonrpc.ServeConn(conn)

		}
	}()

}

func getXgwIPFromRPCConn(conn net.Conn) string {
	ip := conn.RemoteAddr().String()

	ips := strings.Split(ip, ":")

	return ips[0]
}

//ClearAllGreTunnel clear
func (rpc *Xgwka) ClearAllGreTunnel(xgwIP string, reply *Reply) error {

	ueIPArr := tcpamObj.xgwMap[xgwIP]

	for _, ueIP := range *ueIPArr {
		st := tcpamObj.routeMap[ueIP]
		ta := st.ta
		ov := st.ov

		ta.cli.Call("Tcparmp.ReleaseGRETunnel", ta.tcpaIP, &reply.Msg)
		ov.ovsCli.Call("TcpaOvs.ReleaseOvsGRETunnel", ta.tcpaIP, &reply.Msg)

		ta.isIdle = true
		ov.isIdle = true
		ov.ueNum--
		delete(tcpamObj.routeMap, ueIP)
	}

	tcpamObj.xgwMap = make(map[string]*[]string)

	reply.Msg = "succeed"
	return nil
}

//CreateGRETunnel get
func (rpc *Xgwka) CreateGRETunnel(parms Request, reply *Reply) error {

	var st *state
	ueIP := parms.UeIP
	xgwIP := parms.XgwIP

	//查询hss数据库, 判断此ue ip 是否需要tcp加速
	tcpaFlag := QueryTcparByUeIP(ueIP)
	if tcpaFlag != true {
		reply.Msg = fmt.Sprintf("ueIP %s not need ta", ueIP)
		gLoger.WithFields(log.Fields{"ueIP": ueIP}).Infoln("ue not need tcpa")
		return nil
	}

	//gre tunnel exist
	st = tcpamObj.routeMap[ueIP]
	if st.ta != nil && st.ta.isIdle == true {
		reply.ServerIP = st.ta.tcpaIP
		reply.Msg = "succeed"
		return nil
	}

	//获取空闲tcpa代理
	st = getFreeRoute()
	tcpamObj.routeMap[ueIP] = st
	ta := st.ta
	ov := st.ov

	var tcpaParms TcpaRequest
	tcpaParms.UeIP = ueIP
	tcpaParms.TcpaIP = ta.tcpaIP
	tcpaParms.OvsIP = ov.ovsIP

	ta.cli.Call("Tcparmp.CreateGRETunnel", tcpaParms, &reply.Msg)
	if reply.Msg == "succeed" {

		//ovs gre
		ovsRPCCli.Call("TcpaOvs.CreateOvsGRETunnel", ta.tcpaIP, &reply.Msg)
		if reply.Msg != "succeed" {
			ta.cli.Call("Tcparmp.ReleaseGRETunnel", tcpaParms, &reply.Msg)
			ta.isIdle = true
			ov.isIdle = true
			delete(tcpamObj.routeMap, ueIP)
			return nil
		}

		ta.isIdle = false //标识此tcpa被使用
		ov.isIdle = false //标识此ovs 被使用
		ov.ueNum++
		reply.ServerIP = ta.tcpaIP //赋值应答消息tcpa ip
	}

	//add ue to xgw map
	ueIPArr := tcpamObj.xgwMap[xgwIP]
	*ueIPArr = append(*ueIPArr, ueIP)

	gLoger.WithFields(log.Fields{"ovsIP": ov.ovsIP, "ueIP": parms.UeIP, "taIP": ta.tcpaIP, "reply ": reply.Msg}).Infoln("CreateGRETunnel")

	return nil
}

//ReleaseGRETunnel get
func (rpc *Xgwka) ReleaseGRETunnel(parms Request, reply *Reply) error {

	ueIP := parms.UeIP
	st := tcpamObj.routeMap[ueIP]
	if st == nil {
		reply.Msg = "no gre tunnel created"
		return nil
	}

	ta := st.ta
	if ta == nil {
		reply.Msg = fmt.Sprintf("ta %s is not exist", ta.tcpaIP)
		return nil
	}
	ov := st.ov

	ta.cli.Call("Tcparmp.ReleaseGRETunnel", ta.tcpaIP, &reply.Msg)
	if reply.Msg == "succeed" {
		ovsRPCCli.Call("TcpaOvs.ReleaseOvsGRETunnel", ta.tcpaIP, &reply.Msg)
		if reply.Msg == "succeed" {
			ta.isIdle = true
			ov.isIdle = true
			ov.ueNum--
			delete(tcpamObj.routeMap, ueIP)

			gLoger.WithFields(log.Fields{"OvsIP": ov.ovsIP, "ueIP": parms.UeIP, "taIP": ta.tcpaIP, "reply ": reply.Msg}).Infoln("ReleaseGRETunnel tcpa&ovs release succeed")
		} else {

			gLoger.WithFields(log.Fields{"OvsIP": ov.ovsIP, "ueIP": parms.UeIP, "taIP": ta.tcpaIP, "ovs release reply": reply.Msg}).Errorln("ReleaseGRETunnel tcpa release succeed, ovs release failed")

			return nil
		}
	} else {
		gLoger.WithFields(log.Fields{"OvsIP": ov.ovsIP, "ueIP": parms.UeIP, "taIP": ta.tcpaIP, "tcpa release reply": reply.Msg}).Errorln("ReleaseGRETunnel tcpa release failed")
	}

	return nil
}
