package main

import (
	"./Common"
	"./Device"
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

func initAll() error {
	runtime.GOMAXPROCS(runtime.NumCPU())
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
	return err
}

func start() error {
	// TODO
	var err error
	var startTime = time.Now()
	lastTick := startTime.Minute()
	var IMChan, DeviceChan = make(chan error, 1), make(chan error, 1)
	go Data.IMStart(IMChan)
	go Device.LinkHandle(DeviceChan)
	for true {
		select {
		case re := <-IMChan:
			return re
		case re := <-DeviceChan:
			return re
		default:
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
	err = json.Unmarshal(bytes, data)
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
