package main

import (
	"./Client"
	"./Data"
	"./EnvPath"
	"./Log"
	"bufio"
	"encoding/json"
	"io"
	"os"
	"runtime"
	"time"
)

var global = map[string]string{
	"Version": "0.0.8",
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
		return err
	}
	err = readGlobal()
	if err != nil {
		return err
	}
	err = Data.StorageInit()
	if err != nil {
		return err
	}
	auditors = make([]Client.Auditor, 2)
	var t = new(Client.Auditor32375)
	err = auditors[0].Init(t)
	if err != nil {
		return err
	}
	err = auditors[1].Init(&Client.Auditor32376{Password: global["Password"]})
	return err
}

func start() error {
	// TODO
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
