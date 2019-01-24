package Data

import (
	"../Common"
	. "../Log"
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
)

var jsonPath = envpath.GetAppDir()+"/json/YoC.json"
var jsonErrEmpty = errors.New("json is empty")

func IsJsonEmpty(err error) bool {
	if err == jsonErrEmpty {
		return true
	}
	return false
}

func checkJsonDir()error{
	var dir,_ = envpath.GetParentDir(jsonPath)
	return envpath.CheckMakeDir(dir)
}

func JsonRead(device *map[string]*deviceStat) error {
	path, err := jsonPath, checkJsonDir()
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		if os.IsNotExist(err) {
			defer Log.Println("json file not exist")
			return jsonErrEmpty
		} else {
			return err
		}
	}
	var scanner = bufio.NewReader(file)
	bytes, err := scanner.ReadBytes('\n')
	for err != io.EOF {
		if err != nil {
			return err
		}
		var dc dataClass
		err = json.Unmarshal(bytes, &dc)
		if err != nil {
			return err
		}
		(*device)[dc.deviceID] = new(deviceStat)
		(*device)[dc.deviceID].Data = &dc
	}
	defer func() {
		_ = file.Close()
	}()
	return nil
}

func JsonWrite(device *map[string]*deviceStat) error {
	path, err := jsonPath, checkJsonDir()
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 777)
	if err != nil {
		return err
	}
	var writer= bufio.NewWriter(file)
	for _, ds := range *device {
		dc, err := json.Marshal(ds)
		if err != nil {
			return err
		}
		_, err = writer.Write(dc)
		if err != nil {
			return err
		}
		err = writer.WriteByte('\n')
		if err != nil {
			return err
		}
	}
	return nil
}