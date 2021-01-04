package main

import (
	"fmt"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/thinkeridea/go-extend/exnet"
)

var xgwIPMap map[uint]string

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
	UeIP     string `json:"ue_ip"`
	RemoteIP string `json:"remote_ip"`
}

//Reply reply
type Reply struct {
	ServerIP string `json:"server_ip"`
	Msg      string `json:"msg"`
}

func xgwRPCServer() {

	xgwIPMap = make(map[uint]string)

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

			gLoger.WithFields(log.Fields{"addr": xgwIP}).Infoln("xgw sever accept at")

			go jsonrpc.ServeConn(conn)

			id, _ := exnet.IPString2Long(xgwIP)
			if isXgwReConnect(xgwIP) {
				clearAllGreTunnel()
			} else { //xgw 再次上线
				xgwIPMap[id] = xgwIP
				gLoger.WithFields(log.Fields{"xgwIp": xgwIP, "xgwIPMap len": len(xgwIPMap)})
			}
		}
	}()

}

func getXgwIPFromRPCConn(conn net.Conn) string {
	ip := conn.RemoteAddr().String()

	ips := strings.Split(ip, ":")

	return ips[0]
}

func isXgwReConnect(xgwIP string) bool {

	for _, v := range xgwIPMap {
		if v == xgwIP {
			return true
		}
	}

	return false
}

func clearAllGreTunnel() {

	reply := ""

	xgwPool.Range(func(ueIP, tcpaIP interface{}) bool {
		ovsRPCCli.Call("TcpaOvs.ReleaseOvsGRETunnel", tcpaIP, &reply)
		if reply == "succeed" {
			xgwPool.Delete(ueIP)
			gLoger.WithFields(log.Fields{"ueIP": ueIP, "tcpaIP:": tcpaIP, "reply:": reply}).Infoln("clearAllGreTunnel release ovs")
		}
		return true
	})

	var tcpaParms TcpaRequest
	for tcpaIP, ta := range taPool {
		tcpaParms.TcpaIP = tcpaIP
		ta.cli.Call("Tcparmp.ReleaseGRETunnel", tcpaParms, &reply)
		if reply == "succeed" {
			ta.IsIdle = true
		}
		gLoger.WithFields(log.Fields{"taIP": tcpaIP, "reply:": reply}).Infoln("clearAllGreTunnel release tcpa")
	}

}

//CreateGRETunnel get
func (rpc *Xgwka) CreateGRETunnel(parms Request, reply *Reply) error {

	//gre tunnel exist
	if tcpaIP, ok := loadFromXGWPool(parms.UeIP); ok {
		reply.ServerIP = tcpaIP
		reply.Msg = "succeed"
		return nil
	}

	//查询hss数据库, 判断此ue ip 是否需要tcp加速
	tcparFlag := QueryTcparByUeIP(parms.UeIP)
	if tcparFlag != true {
		reply.Msg = fmt.Sprintf("ueIP %s not need ta", parms.UeIP)
		return nil
	}

	//获取空闲tcpa代理
	taIP := getFreeIPFromTaPool()

	ta := getTaFromTaPool(taIP)
	if ta == nil {
		ovsRPCCli.Call("TcpaOvs.AddUeToOvs", parms.UeIP, reply.Msg)
		if reply.Msg == "succeed" {
			reply.Msg = "succeed"
			return nil
		}

		reply.Msg = "no free ta && add ue direct to ovs failed" + "err:" + reply.Msg
		return nil
	}

	ovsIP := GViperCfg.GetString("ovs_ip")

	var tcpaParms TcpaRequest
	tcpaParms.UeIP = parms.UeIP
	tcpaParms.TcpaIP = taIP
	tcpaParms.OvsIP = ovsIP

	ta.cli.Call("Tcparmp.CreateGRETunnel", tcpaParms, &reply.Msg)
	if reply.Msg == "succeed" {
		storeToXGWPool(parms.UeIP, taIP)
		ta.IsIdle = false
		reply.ServerIP = taIP

		//ovs gre
		ovsRPCCli.Call("TcpaOvs.CreateOvsGRETunnel", taIP, &reply.Msg)
		if reply.Msg != "succeed" {
			ta.cli.Call("Tcparmp.ReleaseGRETunnel", tcpaParms, &reply.Msg)
			ta.IsIdle = true

			// if reply.Msg == "connection is shut down" {
			// 	ok := struct{}{}
			// 	ovsCh <- ok
			// }
		}
	}

	gLoger.WithFields(log.Fields{"ovsIP": ovsIP, "ueIP": parms.UeIP, "taIP": taIP, "reply ": reply.Msg}).Infoln("CreateGRETunnel")

	return nil
}

//ReleaseGRETunnel get
func (rpc *Xgwka) ReleaseGRETunnel(parms Request, reply *Reply) error {

	if _, ok := loadFromXGWPool(parms.UeIP); !ok {
		reply.Msg = "no gre tunnel created"
		return nil
	}

	tcpaIP := getTaFromXgwPool(parms.UeIP)
	if tcpaIP == "" {
		reply.Msg = fmt.Sprintf("no found tcpa ip by ue IP:%s", parms.UeIP)
	}

	ta := getTaFromTaPool(tcpaIP)
	if ta == nil {
		reply.Msg = fmt.Sprintf("ta %s is not exist", tcpaIP)
		return nil
	}

	var tcpaParms TcpaRequest
	tcpaParms.UeIP = parms.UeIP
	tcpaParms.TcpaIP = tcpaIP
	tcpaParms.OvsIP = GViperCfg.GetString("ovs_ip")

	ta.cli.Call("Tcparmp.ReleaseGRETunnel", tcpaParms, &reply.Msg)
	if reply.Msg == "succeed" {
		ovsRPCCli.Call("TcpaOvs.ReleaseOvsGRETunnel", tcpaIP, &reply.Msg)
		if reply.Msg == "succeed" {
			ta.IsIdle = true
			deleteIPFromXGWPoolByTaIP(tcpaIP)
			gLoger.WithFields(log.Fields{"OvsIP": tcpaParms.OvsIP, "ueIP": parms.UeIP, "taIP": tcpaIP, "reply ": reply.Msg}).Infoln("ReleaseGRETunnel tcpa&ovs release succeed")
		} else {
			ta.IsIdle = false
			gLoger.WithFields(log.Fields{"OvsIP": tcpaParms.OvsIP, "ueIP": parms.UeIP, "taIP": tcpaIP, "ovs release reply": reply.Msg}).Errorln("ReleaseGRETunnel tcpa release succeed, ovs release failed")

			// if reply.Msg == "connection is shut down" {
			// 	ok := struct{}{}
			// 	ovsCh <- ok
			// }

			return nil
		}
	} else {
		gLoger.WithFields(log.Fields{"OvsIP": tcpaParms.OvsIP, "ueIP": parms.UeIP, "taIP": tcpaIP, "tcpa release reply": reply.Msg}).Errorln("ReleaseGRETunnel tcpa release failed")
	}

	return nil
}
