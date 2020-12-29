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

	exec.Command("ip", "link", "delete", "user").Output()

	exec.Command("ip", "link", "delete", "sat").Output()

	userCmd := exec.Command("ip", "link", "add", "user", "type", "gretap", "remote", "30.254.253.253", "local", "30.254.253.1", "ttl", "255")
	userCmd.Stdout = &out
	userCmd.Stderr = &stderr
	err = userCmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return
	}
	fmt.Println("Result: " + out.String())

	satCmd := exec.Command("ip", "link", "add", "sat", "type", "gretap", "remote", "30.254.253.2", "local", "30.254.253.1", "ttl", "255")
	satCmd.Stdout = &out
	satCmd.Stderr = &stderr
	err = satCmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return
	}
	fmt.Println("Result: " + out.String())

	//go func() {

	//startCmd := exec.Command("/bin/bash", "-C", `/opt/nkt/tcpa/start.sh`)
	startCmd := exec.Command("/opt/nkt/tcpa/start.sh")
	startCmd.Stdout = &out
	startCmd.Stderr = &stderr
	err = startCmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return
	}
	fmt.Println("Result: " + out.String())
	//}()

}
