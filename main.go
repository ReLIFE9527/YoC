package main

import (
	"./Common"
	"./Log"
	"./Manager"
	"encoding/json"
	"os"
	"runtime"
	"time"
)

var global  = map[string]string{
	"Version":"0.0.2"}

func initAll() error {
	runtime.GOMAXPROCS(runtime.NumCPU())
	var err= YoCLog.LogInit()
	if err != nil {
		return err
	}
	filePath := envpath.GetAppDir()
	filePath += "/YoC.info"
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	version, err := json.Marshal(global)
	_, err = file.Write(version)
	if err != nil {
		return err
	}
	err = Data.IMInit()
	return err
}

func start() error {
	// TODO
	var err error
	var startTime =time.Now()
	lastTick := startTime.Minute()
	var IMChan =make(chan error,1)
	go Data.IMStart(&IMChan)
	//go Data.IMDeviceLogin("aaaa")
	for true {
		select {
		case re := <-IMChan:
			return re
		default:
			t := time.Now()
			if t.Second() == startTime.Second() && lastTick-t.Minute()%10==0 {
				YoCLog.Log.Println("minute tick ", t)
				lastTick = t.Minute()
			}
		}
	}
	return err
}

func exit(ec error) {
	err := Data.IMShutDown()
	if err != nil {
		YoCLog.Log.Println(err)
	}
	YoCLog.LogExit(ec)
}

func main() {
	var ret= initAll()
	if ret!=nil {
		exit(ret)
	}
	ret = start()
	defer exit(ret)
}