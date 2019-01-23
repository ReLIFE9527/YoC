package main

import (
	"./Log"
)

func initAll() error {
	var err = YoCLog.LogInit()
	if err!=nil{
		return err
	}else{
		return nil
	}
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