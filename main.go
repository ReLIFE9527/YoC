package main

import (
	"./Client"
	"./Device"
	"./EnvPath"
	"./Log"
	"./Manager"
	"bufio"
	"encoding/json"
	"io"
	"os"
	"runtime"
	"time"
)

var global = map[string]string{
	"Version": "0.0.6",
}

var moduleChannelProperty = map[string]int{
	"IM":     1,
	"Device": 1,
	"Client": 1,
}
var moduleChannel map[string]chan error

func initAll() error {
	runtime.GOMAXPROCS(runtime.NumCPU())
	go initChannel()
	var err = YoCLog.LogInit()
	if err != nil {
		return err
	}
	err = readGlobal()
	if err != nil {
		return err
	}
	err = Data.IMInit()
	if err != nil {
		return err
	}
	err = Device.LinkInit()
	if err != nil {
		return err
	}
	err = Client.LinkInit(global["Password"])
	return err
}

func start() error {
	// TODO
	var err error
	var startTime = time.Now()
	lastTick := startTime.Minute()
	go Data.IMStart(moduleChannel["IM"])
	go Device.LinkHandle(moduleChannel["Device"])
	go Client.LinkHandle(moduleChannel["Client"])
	for true {
		select {
		case re := <-moduleChannel["IM"]:
			return re
		case re := <-moduleChannel["Device"]:
			return re
		case re := <-moduleChannel["Client"]:
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
	err := Data.IMShutDown()
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

func readGlobal() (err error) {
	var data = make(map[string]string)
	filePath := envpath.GetAppDir()
	filePath += "/YoC.info"
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDONLY, os.ModePerm)
	scanner := bufio.NewReader(file)
	bytes, err := scanner.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return err
	}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return err
	}
	if ps, ok := data["Password"]; !ok || ps == "" {
		data["Password"] = "YoCProject"
	}
	if data["Version"] == global["Version"] {
		global = data
	} else {
		data["Version"] = global["Version"]
		global = data
		file, err = os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
		if err != nil {
			return err
		}
		globalInfo, err := json.Marshal(global)
		if err != nil {
			return err
		}
		_, err = file.Write(globalInfo)
		if err != nil {
			return err
		}
	}
	return err
}

func initChannel() {
	moduleChannel = make(map[string]chan error)
	for name, buf := range moduleChannelProperty {
		moduleChannel[name] = make(chan error, buf)
	}
}
