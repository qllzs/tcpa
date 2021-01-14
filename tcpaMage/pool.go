package main

import (
	"net/rpc"

	log "github.com/sirupsen/logrus"
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
	tcpaMap  map[string]*tcpa     //taIP  -- ta
	ovsMap   map[string]*ovs      //ovsIP  -- ovs
}

var tcpamObj *tcpam

func init() {
	GViperCfg.SetDefault("max_ue_num", 10)
	MaxUeNum = GViperCfg.GetInt("max_ue_num")
	gLoger.WithFields(log.Fields{"MaxUeNum": MaxUeNum}).Infoln("ovs init")
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
				if ov.isIdle == true && ov.ovsCli != nil { //not used -- cli ok  -- ueNum ok
					st.ta = ta
					st.ov = ov

					return &st
				}

			}
		}
	}

	return nil
}
