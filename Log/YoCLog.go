package YoCLog

import (
	"log"
	"os"
	"time"
)

import 	"../Common"

var Log *log.Logger
var logFile *os.File

func LogInit() {
	logFile =openLog()
	Log = log.New(logFile,"",log.LstdFlags|log.Lshortfile|log.Ltime|log.Llongfile)
	Log.Println("---------------Log Start---------------")
}

func LogExit(ec int64) {
	Log.Printf("Exit with code %d\n",ec)
	Log.Println("---------------Log End----------------")
	_=logFile.Close()
}

func openLog() *os.File{
	var path= envpath.GetLogPath("YoC")
	stat,err := os.Stat(path)
	log.Println(stat.Size())
	if stat.Size()>0x80000 {
		err = os.Rename(path, path+"."+time.Now().Format("2006-1-2 15-04-05"))
		if err != nil {
			log.Println("failed to rename log file at " + path)
			log.Fatal(err)
		}
	}
	file,err:= os.OpenFile(path, os.O_APPEND|os.O_CREATE, 777)
	if err != nil {
		log.Println("failed to load log file at: " + path)
		log.Fatal(err)
	}
	return file
}