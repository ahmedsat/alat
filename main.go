package main

import (
	"os"
	"runtime"
)

type Mod string

const (
	ModClient Mod = "client"
	ModServer Mod = "server"
)

func init() {
	runtime.LockOSThread()
}

func main() {

	if len(os.Args) < 2 {
		server := Server{}
		server.Run()
	} else {
		Client := Client{}
		Client.Run(os.Args[1:]...)
	}
}
