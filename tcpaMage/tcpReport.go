package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/thinkeridea/go-extend/exnet"
)

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

var beatCh chan struct{}

func report() {
	beatCh = make(chan struct{})

	tcpAddr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:50051")
	if err != nil {
		fmt.Println(err)
	}

	tcpListener, _ := net.ListenTCP("tcp", tcpAddr)

	gLoger.WithFields(log.Fields{"tcpAddr": tcpAddr}).Infoln("report listen at")

	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			fmt.Println(err)
			continue
		}
		gLoger.WithFields(log.Fields{"ip:port": tcpConn.RemoteAddr().String()}).Infoln("accept ta client")

		go func() {
			data := make([]byte, 1024)
			for {
				total, err := tcpConn.Read(data)
				if err != nil {
					gLoger.WithFields(log.Fields{"err:": err.Error(), "ip:port": tcpConn.RemoteAddr().String()}).Errorln("failed to read tcp msg")
					return
				}

				err = decodeReport(tcpConn, total, data)
				if err != nil {
					gLoger.WithFields(log.Fields{"ip:port": tcpConn.RemoteAddr().String()}).Errorln("decodeReport" + err.Error())
					return
				}
			}
		}()
	}
}

func decodeReport(tcpConn net.Conn, total int, data []byte) error {

	var req ReportMsg
	json.Unmarshal(data[:total], &req)

	switch req.Type {
	case MsgTypeConnect:
		{
			//taIP, _ := exnet.Long2IP(req.IP)
			taIP, _ := exnet.Long2IPString(req.IP)

			ips := strings.Split(taIP, ":")
			taServerIP := ips[0]

			tcpamObj.ueNum++

			var ta tcpa
			ta.tcpaIP = taServerIP
			tcpamObj.tcpaNum++
			tcpamObj.tcpaMap[taServerIP] = &ta

			err := tcpaRPCClient(taServerIP)
			if err != nil {
				tcpamObj.tcpaNum--
				delete(tcpamObj.tcpaMap, taServerIP)
				return err
			}

			go hearBeat(taServerIP)

			//应答连接report
			tcpConn.Write([]byte("succeed"))

			return nil
		}

	case MsgTypeReport:
		switch req.State {
		case true:
			ok := struct{}{}
			beatCh <- ok
		case false:
			// taIP, _ := exnet.Long2IP(req.IP)
			// deleteIPFromXGWPoolByTaIP(taIP.String())
		}

	}

	return nil
}

func hearBeat(taServerIP string) {

	for {
		ticker := time.NewTicker(time.Second * 11)

		select {
		case <-beatCh:
			ticker.Stop()
			continue
		case <-ticker.C:
			tcpamObj.tcpaNum--
			delete(tcpamObj.tcpaMap, taServerIP)
			gLoger.WithFields(log.Fields{"ip:port": taServerIP}).Errorln("no tcpa hearBeat, dead")
			return
		}
	}

}
