package main

import (
	"encoding/json"
	"fmt"
	"net"
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

			tcpamObj.ueNum++

			var ta tcpa
			ta.tcpaIP = taIP
			tcpamObj.tcpaNum++
			tcpamObj.tcpaMap[taIP] = &ta

			err := tcpaRPCClient(taIP)
			if err != nil {
				delete(tcpamObj.tcpaMap, taIP)
				return err
			}

			go hearBeat(taIP)

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

func hearBeat(taIP string) {

	for {
		ticker := time.NewTicker(time.Second * 11)

		select {
		case <-beatCh:
			ticker.Stop()
			continue
		case <-ticker.C:
			delete(tcpamObj.tcpaMap, taIP)
			gLoger.WithFields(log.Fields{"ip:port": taIP}).Errorln("no tcpa hearBeat, dead")
			return
		}
	}

}
