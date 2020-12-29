package main

func main() {

	go report()
	go ovsRPCClient()

	xgwRPCServer()

	for {

	}
}
