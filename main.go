package main

import (
	"./Common"
	"./Log"
	"encoding/json"
	"os"
)

var global  = map[string]string{
	"Version":"0.0.1"}

func initAll() error {
	var err = YoCLog.LogInit()
	if err != nil {
		return err
	}
	filePath := envpath.GetAppDir()
	filePath += "/YoC.info"
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return nil
	}
	version,err:=json.Marshal(global)
	_, err = file.Write(version)
	return err
}

func start() error {
	// TODO
	return nil
}

func exit(ec error) {
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