package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func InitMainParam(args []string) {

	if len(args) < 2 {
		return
	}

	arg := args[1]
	switch arg {
	case "-h":
		fmt.Printf("This is help")
		os.Exit(0)
	case "-v":
		fmt.Println("Version :" + gVersion)
		os.Exit(0)
	default:
		fmt.Printf("Unknown argument:%s", arg)
	}

}

func InitLog() {
	logs.LoadConfiguration("../etc/log4go.xml")
	logs.Info("load log4go!")
}

func InitSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM) //,syscall.SIGUSR1
	go func() {
		s := <-signals
		logs.Warn("singnal:" + s.String())
		gRun = 0

		//1秒后强制退出,避免挂死
		//time.Sleep(time.Second * 1)
		//os.Exit(0)
	}()
}

func timer() {

}

func main() {

	gRun = 1
	//gTaskMovieMap = make(map[string]MovieAdd)

	InitLog()
	InitSignals()
	LoadConfig(gConfigPath)
	InitMainParam(os.Args)
	InitHttpService(gConfig.Host, gConfig.Port)

	for {

		if gRun == 0 {
			logs.Info("singnal exit !")
			return
		}

		gNow = time.Now().Unix()

		timer()
		time.Sleep(time.Second * 1)
	}

}
