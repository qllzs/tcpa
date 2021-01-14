package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/thinkeridea/go-extend/exnet"
)

var tcpConn net.Conn

//Msg type
const (
	_ int = iota
	MsgTypeConnect
	MsgTypeReport
)

//ReportMsg report
type ReportMsg struct {
	Type  int  `json:"type"`
	IP    uint `json:"ip"`
	State bool `json:"state"`
}

func connect() {
	var err error

	for {

		tcpaMageIP := GViperCfg.GetString("tcpa_manage_ip")
		tcpConn, err = net.Dial("tcp", tcpaMageIP+":50051")
		if err != nil {
			log.Errorln("Can't dial: ", err)
			time.Sleep(time.Second * 2)
			continue
		}
		log.Infoln("Dail at:", tcpConn.RemoteAddr().String())

		err := connectReport()
		if err != nil {
			log.Errorln("connect:", err)
			continue
		}

		for {
			ticker := time.NewTicker(time.Second * 3)

			<-ticker.C

			err := report()
			if err != nil {
				break
			}
		}

	}

}

func connectReport() error {

	_, err := tcpConn.Write(encodeReportMsg(MsgTypeConnect, true))
	if err != nil {
		return err
	}

	tcpConn.SetReadDeadline(time.Now().Add(time.Second * 2))

	data := make([]byte, 10)
	n, err := tcpConn.Read(data)
	if err != nil {
		return err
	}

	if strings.Compare(string(data[:n]), "succeed") != 0 {
		return errors.New("ta manage response failed")
	}

	return nil
}

func report() error {

	_, err := tcpConn.Write(encodeReportMsg(MsgTypeReport, true))
	if err != nil {
		log.Errorln("Report:", err)
		return err
	}
	return nil
}

func encodeReportMsg(t int, s bool) []byte {

	var msg ReportMsg
	var err error

	switch t {
	case MsgTypeConnect:
		msg.Type = MsgTypeConnect
		msg.IP, err = exnet.IP2Long(externalIP("ens160"))
		if err != nil {
			log.Errorln("get ens160 ip failed:", err)
			return nil
		}
	case MsgTypeReport:
		msg.Type = MsgTypeReport
	}

	msg.State = true

	data, err := json.Marshal(msg)
	if err != nil {
		log.Errorln("marsh report msg failed:", err)
		return []byte{}
	}
	return data
}

func externalIP(name string) net.IP {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil
	}

	if iface.Flags&net.FlagUp == 0 { // interface down
		return nil
	}

	addres, err := iface.Addrs()
	if err != nil {
		fmt.Println("err", err)
	}
	for _, addr := range addres {
		ip := getIPFromAddr(addr)
		if ip != nil {
			return ip
		}
	}

	return nil
}

func getIPFromAddr(addr net.Addr) net.IP {

	var ip net.IP

	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil {
		return nil
	}

	ip = ip.To4()
	if ip == nil {
		return nil // not an ipv4 address
	}

	return ip
}
