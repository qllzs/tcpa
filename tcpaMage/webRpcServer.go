package main

import (
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"

	log "github.com/sirupsen/logrus"
)

//Web web
type Web struct {
}

//TcpaList tcpa list
type TcpaList struct {
	TaIP   string `json:"tcpa_ip"`
	IsIdle bool   `json:"is_idle"`
}

//TcpaState tcpa
type TcpaState struct {
	TaTotalNum int `json:"tcpa_total_num"`
	TaUsedNum  int `json:"tcpa_used_num"`
	TaFreeNum  int `json:"tcpa_free_num"`
	TaList     []TcpaList
}

//OvsList ovs list
type OvsList struct {
	OvsIP  string `json:"ovs_ip"`
	UeNum  int    `json:"ue_num"`
	IsIdle bool   `json:"is_idle"`
}

//OvsState ovs
type OvsState struct {
	OvTotalNum int `json:"ovs_total_num"`
	OvUsedNum  int `json:"ovs_used_num"`
	OvFreeNum  int `json:"ovs_free_num"`
	OvUeNum    int `json:"ovs_ue_num"`
	OvList     []OvsList
}

//RouteList route list
type RouteList struct {
	UeIP   string `json:"ue_ip"`
	OvsIP  string `json:"ovs_ip"`
	TcpaIP string `josn:"tcpa_ip"`
}

//RouteState route
type RouteState struct {
	RtTotalNum int `json:"route_total_num"`
	RtList     []RouteList
}

//WebReply reply
type WebReply struct {
	Ta TcpaState  `json:"ta"`
	Ov OvsState   `json:"ov"`
	Rt RouteState `json:"rt"`
}

func init() {

	rpc.Register(new(Web))

	lis, err := net.Listen("tcp", ":50055")
	if err != nil {
		gLoger.WithFields(log.Fields{"err": err.Error()}).Errorln("webRPCServer xgw rpc sever listen failed")
		return
	}
	gLoger.WithFields(log.Fields{"ip": lis.Addr().String()}).Errorln("webRPCServer xgw rpc sever listen at")

	go func() {
		for {
			conn, err := lis.Accept()
			if err != nil {
				gLoger.WithFields(log.Fields{"err": err.Error()}).Errorln("web sever accept failed")
				continue
			}
			gLoger.WithFields(log.Fields{"web ip ": conn.RemoteAddr().String()}).Infoln("web sever accept at")

			go jsonrpc.ServeConn(conn)

		}
	}()

}

//GetTcpamState get
func (rpc *Web) GetTcpamState(parm string, reply *WebReply) error {

	*reply = getallState()

	return nil
}

func getTcpaState() TcpaState {

	var taState TcpaState
	taState.TaList = make([]TcpaList, 0)

	for taIP, ta := range tcpamObj.tcpaMap {

		var taList TcpaList
		if ta.isIdle == true {
			taState.TaFreeNum++
		} else {
			taState.TaUsedNum++
		}

		taList.TaIP = taIP
		taList.IsIdle = ta.isIdle

		taState.TaTotalNum++
		taState.TaList = append(taState.TaList, taList)
	}

	return taState
}

func getOvsState() OvsState {

	var ovState OvsState

	ovState.OvList = make([]OvsList, 0)

	for ovIP, ov := range tcpamObj.ovsMap {
		var ovList OvsList
		if ov.isIdle == false { //ovs满负载(10),标识为不可用
			ovState.OvUsedNum++
		} else {
			ovState.OvFreeNum++
		}

		ovList.OvsIP = ovIP
		ovList.IsIdle = ov.isIdle
		ovList.UeNum = ov.ueNum

		ovState.OvTotalNum++
		ovState.OvUeNum += ov.ueNum
		ovState.OvList = append(ovState.OvList, ovList)
	}

	gLoger.WithFields(log.Fields{"total": ovState.OvTotalNum,
		"used": ovState.OvUsedNum,
		"free": ovState.OvFreeNum,
		"list": ovState.OvList}).Infoln("getOvsState")

	return ovState
}

func getRouteState() RouteState {

	var rtState RouteState
	rtState.RtList = make([]RouteList, 0)
	for ueIP, st := range tcpamObj.routeMap {

		var rt RouteList
		rt.OvsIP = st.ov.ovsIP
		rt.TcpaIP = st.ta.tcpaIP
		rt.UeIP = ueIP
		rtState.RtTotalNum++
		rtState.RtList = append(rtState.RtList, rt)
	}

	return rtState
}

func getallState() WebReply {
	var state WebReply

	state.Ov = getOvsState()
	state.Ta = getTcpaState()
	state.Rt = getRouteState()

	// rs, err := json.Marshal(state)
	// if err != nil {
	// 	return []byte("getOvsState make json error")
	// }

	return state
}
