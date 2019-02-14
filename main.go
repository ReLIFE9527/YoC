package main

import (
	"./Client"
	"./Data"
	"./Log"
	"log"
	"runtime"
	"time"
)

var global = map[string]string{
	"Version": "0.2.0",
}

var moduleChannelProperty = map[string]int{
	"repository": 1,
	"port:32375": 1,
	"port:32376": 1,
}
var moduleChannel map[string]chan error

var auditors []Client.Auditor

func initAll() error {
	runtime.GOMAXPROCS(runtime.NumCPU())
	go initChannel()
	var err = YoCLog.LogInit()
	if err != nil {
		log.Println(err)
		return err
	}
	err = Data.ReadGlobal(global)
	if err != nil {
		log.Println(err)
		return err
	}
	err = Data.StorageInit()
	if err != nil {
		log.Println(err)
		return err
	}
	auditors = make([]Client.Auditor, 2)
	var t = new(Client.Auditor32375)
	err = auditors[0].Init(t)
	if err != nil {
		log.Println(err)
		return err
	}
	err = auditors[1].Init(&Client.Auditor32376{Password: global["Password"]})
	return err
}

func start() error {
	var err error
	var startTime = time.Now()
	lastTick := startTime.Minute()
	go Data.StorageStart(moduleChannel["repository"])
	go auditors[0].Listen(moduleChannel["port:32375"])
	go auditors[1].Listen(moduleChannel["port:32376"])
	for true {
		select {
		case re := <-moduleChannel["repository"]:
			return re
		case re := <-moduleChannel["port:32375"]:
			return re
		case re := <-moduleChannel["port:32376"]:
			return re
		default:
			<-time.After(time.Second)
			t := time.Now()
			if t.Second() == startTime.Second() && lastTick-t.Minute()%10 == 0 {
				YoCLog.Log.Println("minute tick ", t)
				lastTick = t.Minute()
			}
		}
	}
	return err
}

func exit(ec error) {
	err := Data.StorageShutDown()
	if err != nil {
		YoCLog.Log.Println(err)
	}
	YoCLog.LogExit(ec)
}

func main() {
	var ret = initAll()
	if ret != nil {
		exit(ret)
		return
	}
	ret = start()
	defer exit(ret)
}

func initChannel() {
	moduleChannel = make(map[string]chan error)
	for name, buf := range moduleChannelProperty {
		moduleChannel[name] = make(chan error, buf)
	}
}
