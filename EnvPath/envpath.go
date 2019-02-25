package envpath

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var appDir = func() string {
	switch runtime.GOOS {
	case "windows":
		path, err := filepath.Abs("./")
		if err != nil {
			log.Fatal(err)
		}
		return filepath.ToSlash(path)
	case "linux":
		path, err := filepath.Abs("./")
		if err != nil {
			log.Fatal(err)
		}
		path, err = GetParentDir(path)
		if err != nil {
			log.Fatal(err)
		}
		return path
	default:
		log.Fatal("operation system type err: " + runtime.GOOS)
		return ""
	}
}()

func GetAppDir() string {
	return appDir
}

func GetParentDir(srcPath string) (dstPath string, err error) {
	if srcPath[len(srcPath)-1] == '/' {
		srcPath = srcPath[:len(srcPath)-1]
	}
	var index = strings.LastIndexByte(srcPath, '/')
	dstPath = srcPath[:index]
	_, err = os.Stat(dstPath)
	return dstPath, err
}

func CheckMakeDir(dir string) error {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err = os.Mkdir(dir, os.ModePerm)
	}
	return err
}

func GetSubPath(srcPath string, subDir string) (dstPath string, err error) {
	if srcPath[len(srcPath)-1] == '/' {
		srcPath = srcPath[:len(srcPath)-1]
	}
	dstPath = srcPath + "/" + subDir
	_, err = ioutil.ReadDir(dstPath)
	if err != nil {
		return "", err
	}
	return dstPath, nil
}

func GetSubFile(srcPath string, subFile string) (dstPath string, err error) {
	if srcPath[len(srcPath)-1] == '/' {
		srcPath = srcPath[:len(srcPath)-1]
	}
	dstPath = srcPath + "/" + subFile
	_, err = ioutil.ReadFile(dstPath)
	if err != nil {
		return "", err
	}
	return dstPath, nil
}
