package Data

import (
	"../EnvPath"
	. "../Log"
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
)

var jsonPath = envpath.GetAppDir() + "/json/YoC.json"
var jsonErrEmpty = errors.New("json is empty")

func IsJsonEmpty(err error) bool {
	if err == jsonErrEmpty {
		return true
	}
	return false
}

func checkJsonDir() error {
	var dir, _ = envpath.GetParentDir(jsonPath)
	return envpath.CheckMakeDir(dir)
}

func JsonRead(device *map[string]*stat) error {
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
		var dc repository
		err = json.Unmarshal(bytes, &dc)
		if err != nil {
			return err
		}
		(*device)[dc.id] = new(stat)
		(*device)[dc.id].Data = &dc
		bytes, err = scanner.ReadBytes('\n')
	}
	defer func() {
		_ = file.Close()
	}()
	return nil
}

func JsonWrite(device *map[string]*stat) error {
	path, err := jsonPath, checkJsonDir()
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	defer func() {
		_ = file.Close()
	}()
	if err != nil {
		return err
	}
	for _, ds := range *device {
		if ds == nil {
			continue
		}
		dc, err := json.Marshal(ds.Data)
		if err != nil {
			return err
		}
		dc = append(dc, '\n')
		_, err = file.Write(dc)
		if err != nil {
			return err
		}
	}
	return nil
}
