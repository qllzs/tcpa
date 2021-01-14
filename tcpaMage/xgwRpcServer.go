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
	UeIP  string `json:"ue_ip"`
	XgwIP string `json:"xgw_ip"`
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

			if tcpamObj.xgwMap[xgwIP] == nil {
				tcpamObj.xgwNum++
				ueIPArr := []string{}
				tcpamObj.xgwMap[xgwIP] = &ueIPArr
			}

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
func (rpc *Xgwka) ClearAllGreTunnel(parms Request, reply *Reply) error {

	xgwIP := parms.XgwIP

	ueIPArr := tcpamObj.xgwMap[xgwIP]

	if len(*ueIPArr) != 0 {

		gLoger.WithFields(log.Fields{"ue num": len(*ueIPArr), "xgw ip": xgwIP}).Infoln("ClearAllGreTunnel ue number at xgw")
		for _, ueIP := range *ueIPArr {
			st := tcpamObj.routeMap[ueIP]
			ta := st.ta
			ov := st.ov

			ta.cli.Call("Tcparmp.ReleaseGRETunnel", ta.tcpaIP, &reply.Msg)
			ov.ovsCli.Call("TcpaOvs.ReleaseOvsGRETunnel", ta.tcpaIP, &reply.Msg)

			ta.isIdle = true
			ov.ueNum--
			if ov.isIdle == false && ov.ueNum < MaxUeNum {
				ov.isIdle = true
			}

			//删除ue 使用的tcpa ovs的 gre 通道
			delete(tcpamObj.routeMap, ueIP)
			gLoger.WithFields(log.Fields{"ue ip": ueIP, "tcpa ip": ta.tcpaIP, "ovs ip": ov.ovsIP}).Infoln("ClearAllGreTunnel delete route ")
		}
		//清空xgw ip 对应的ue
		*ueIPArr = []string{}
	}

	gLoger.WithFields(log.Fields{"xgw ip": xgwIP}).Infoln("ClearAllGreTunnel clear all gre tunnel by xgw")
	reply.Msg = "succeed"
	return nil
}

//CreateGRETunnel get
func (rpc *Xgwka) CreateGRETunnel(parms Request, reply *Reply) error {

	var st *state
	ueIP := parms.UeIP
	xgwIP := parms.XgwIP

	ueIPArr := tcpamObj.xgwMap[xgwIP]
	if ueIPArr == nil {
		reply.Msg = "no xgw ip connect, xgw ip failed"
		return nil
	}

	//查询hss数据库, 判断此ue ip 是否需要tcp加速
	tcpaFlag := QueryTcparByUeIP(ueIP)
	if tcpaFlag != true {
		reply.Msg = fmt.Sprintf("ueIP %s not need ta", ueIP)
		gLoger.WithFields(log.Fields{"ueIP": ueIP}).Infoln("ue not need tcpa")
		return nil
	}

	//gre tunnel exist
	st = tcpamObj.routeMap[ueIP]
	if st != nil && st.ta != nil && st.ta.isIdle == true {
		reply.ServerIP = st.ta.tcpaIP
		reply.Msg = "succeed"
		return nil
	}

	//获取空闲tcpa代理
	st = getFreeRoute()
	if st == nil {
		reply.Msg = "no gre tunnel"
		return nil
	}

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
		ov.ovsCli.Call("TcpaOvs.CreateOvsGRETunnel", ta.tcpaIP, &reply.Msg)
		if reply.Msg != "succeed" {
			ta.cli.Call("Tcparmp.ReleaseGRETunnel", tcpaParms, &reply.Msg)
			delete(tcpamObj.routeMap, ueIP)
			return nil
		}

		ta.isIdle = false
		ov.ueNum++
		if ov.isIdle == true && ov.ueNum >= MaxUeNum {
			ov.isIdle = false
		}

		reply.ServerIP = ta.tcpaIP //赋值应答消息tcpa ip
	}

	//add ue to xgw map
	*ueIPArr = append(*ueIPArr, ueIP)

	gLoger.WithFields(log.Fields{"ovsIP": ov.ovsIP, "ueIP": parms.UeIP, "taIP": ta.tcpaIP, "reply ": reply.Msg}).Infoln("CreateGRETunnel")

	return nil
}

//ReleaseGRETunnel get
func (rpc *Xgwka) ReleaseGRETunnel(parms Request, reply *Reply) error {

	ueIP := parms.UeIP

	st := tcpamObj.routeMap[ueIP]
	if st == nil || st.ta == nil || st.ov == nil {
		reply.Msg = "no gre tunnel created"
		return nil
	}

	ta := st.ta
	ov := st.ov

	ta.cli.Call("Tcparmp.ReleaseGRETunnel", ta.tcpaIP, &reply.Msg)
	if reply.Msg == "succeed" {
		ov.ovsCli.Call("TcpaOvs.ReleaseOvsGRETunnel", ta.tcpaIP, &reply.Msg)
		if reply.Msg == "succeed" {
			ta.isIdle = true
			ov.ueNum--
			if ov.isIdle == false && ov.ueNum < MaxUeNum {
				ov.isIdle = true
			}
			delete(tcpamObj.routeMap, ueIP)

			gLoger.WithFields(log.Fields{"OvsIP": ov.ovsIP, "ueIP": ueIP, "taIP": ta.tcpaIP, "reply ": reply.Msg}).Infoln("ReleaseGRETunnel tcpa&ovs release succeed")
		} else {

			gLoger.WithFields(log.Fields{"OvsIP": ov.ovsIP, "ueIP": ueIP, "taIP": ta.tcpaIP, "ovs release reply": reply.Msg}).Errorln("ReleaseGRETunnel tcpa release succeed, ovs release failed")

			return nil
		}
	} else {
		gLoger.WithFields(log.Fields{"OvsIP": ov.ovsIP, "ueIP": ueIP, "taIP": ta.tcpaIP, "tcpa release reply": reply.Msg}).Errorln("ReleaseGRETunnel tcpa release failed")
	}

	return nil
}
