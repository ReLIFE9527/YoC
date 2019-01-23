package envpath

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"strings"
)

var osType = runtime.GOOS

func GetPath(path string) string {
	var ret string
	var err error
	switch osType {
	case "windows":
		path = strings.Replace(path, "/", "\\", -1)
		ret, err = filepath.Abs("./")
		if err != nil {
			log.Fatal("Can't get path")
		}
		if path[0] != '\\' {
			path = "\\" + path
		}
		ret += path
	case "linux":
		if path[0] == '/' {
			path = path[1:]
		}
		ret = "/root/" + path
	default:
	}
	fmt.Println("Log path now is " + ret)
	return ret
}

func GetLogPath(project string)string {
	return GetPath("logs/" + project + ".log")
}

func GetDBPath(project string) string{
	return GetPath("app/"+project+"db.json")
}