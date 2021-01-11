package main

import (
	"net/rpc"
)

var MaxUeNum int

type state struct {
	ov *ovs
	ta *tcpa
}

type tcpa struct {
	isIdle bool //idle : true, used: false
	cli    *rpc.Client
	tcpaIP string
}

type ovs struct {
	isIdle bool
	ueNum  int
	ovsCli *rpc.Client
	ovsIP  string
	rpcIP  string
}

type tcpam struct {
	xgwNum   int                  //xgw number
	ueNum    int                  //ue number
	ovsNum   int                  //ovs number
	tcpaNum  int                  //tcpa total num
	xgwMap   map[string]*[]string //xgwIP - ueIP
	routeMap map[string]*state    //ueIP - state
	tcpaMap  map[string]*tcpa
	ovsMap   map[string]*ovs
}

var tcpamObj *tcpam

func init() {
	GViperCfg.SetDefault("max_ue_num", 10)
	MaxUeNum = GViperCfg.GetInt("max_ue_num")
	tcpamObj = getNewTcpam()
}

func getNewTcpam() *tcpam {

	var tcpamObj tcpam

	tcpamObj.xgwNum = 0
	tcpamObj.ueNum = 0
	tcpamObj.ovsNum = 0
	tcpamObj.tcpaNum = 0

	tcpamObj.xgwMap = make(map[string]*[]string)
	tcpamObj.routeMap = make(map[string]*state)
	tcpamObj.tcpaMap = make(map[string]*tcpa)
	tcpamObj.ovsMap = make(map[string]*ovs)

	return &tcpamObj
}

func getFreeRoute() *state {

	var st state

	for _, ta := range tcpamObj.tcpaMap {
		if ta.isIdle == true && ta.cli != nil { //not used
			for _, ov := range tcpamObj.ovsMap {
				if ov.isIdle == true && ov.ovsCli != nil && ov.ueNum <= MaxUeNum { //not used -- cli ok  -- ueNum ok
					ta.isIdle = false
					st.ta = ta
					ov.isIdle = false
					st.ov = ov
					return &st
				}
			}
		}
	}

	return nil
}
