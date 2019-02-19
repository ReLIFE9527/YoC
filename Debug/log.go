package YoCLog

import (
	"log"
	"os"
	"time"
)

import "../EnvPath"

var DebugLogger *log.Logger
var logFile *os.File

func LogInit() error {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Ltime)
	logFile = openLog()
	DebugLogger = log.New(logFile, "", log.LstdFlags|log.Lshortfile|log.Ltime)
	DebugLogger.Println("---------------Log Start---------------")
	return nil
}

func LogExit(ec error) {
	defer func() {
		_ = logFile.Close()
	}()
	if ec != nil {
		DebugLogger.Println("exit with error", ec)
	} else {
		DebugLogger.Println("normal exit")
	}
	DebugLogger.Println("---------------Log End----------------")
}

func openLog() *os.File {
	var filePath = envpath.GetAppDir() + "/logs/YoC.log"
	stat, err := os.Stat(filePath)
	if err != nil {
		log.Println(err)
	} else {
		//log.Println("log size now is ",stat.Size())
		if stat.Size() > 0x80000 {
			err = os.Rename(filePath, filePath+"."+time.Now().Format("2006-1-2 15-04-05"))
			if err != nil {
				log.Println("failed to rename log file at " + filePath)
				log.Fatal(err)
			}
		}
	}
	dir, _ := envpath.GetParentDir(filePath)
	err = envpath.CheckMakeDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Println("failed to load log file at: " + filePath)
		log.Fatal(err)
	}
	_, err = file.WriteString("\n")
	if err != nil {
		log.Println(err)
	}
	return file
}
