package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

func main() {
	var err error
	var out bytes.Buffer
	var stderr bytes.Buffer

	//ovs-vsctl del-port 30.254.253.1
	exec.Command("ovs-vsctl", "del-port", "30.254.253.1").Output()

	// //ovs-vsctl add-port tcpa_ovs_br 30.254.253.1 -- set interface 30.254.253.1 type=gre options:remote_ip=30.254.253.1
	cmd := exec.Command("ovs-vsctl", "add-port", "tcpa_ovs_br", "30.254.253.1", "--", "set", "interface", "30.254.253.1", "type=gre", "options:remote_ip="+"30.254.253.1")
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("ovs:" + fmt.Sprint(err) + ": " + stderr.String())
		return
	}
	fmt.Println("ovs:" + "Result: " + out.String())

}
