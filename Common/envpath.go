package envpath

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var appDir = func()string{
	switch runtime.GOOS {
	case "windows":
		path,err :=filepath.Abs("./")
		if err!=nil{
			log.Fatal(err)
		}
		return filepath.ToSlash(path)
	case "linux":
		return "/root/"
	default:
		log.Fatal("opration system type err: "+runtime.GOOS)
		return ""
	}
}

func GetAppDir()string {
	return appDir()
}

func GetParentDir(srcPath string) (dstPath string,err error) {
	if srcPath[len(srcPath)-1] == '/' {
		srcPath = srcPath[:len(srcPath)-1]
	}
	var index = strings.LastIndexByte(srcPath, '/')
	dstPath = srcPath[:index+1]
	_, err = os.Stat(dstPath)
	return dstPath, err
}

func CheckMakeDir(dir string) error {
	_,err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(dir, os.ModePerm)
	}
	return err
}