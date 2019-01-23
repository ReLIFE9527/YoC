package main

import "./Log"

func initAll() {
	YoCLog.LogInit()
}

func start() int64 {
	// TODO
	return 0
}

func exit(ec int64) {
	YoCLog.LogExit(ec)
}

func main() {
	initAll()
	var ret= start()
	exit(ret)
}