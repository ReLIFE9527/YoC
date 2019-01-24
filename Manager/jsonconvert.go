package Data

import (
	"../Common"
	. "../Log"
	"bufio"
	"encoding/json"
	"io"
	"os"
)

var jsonPath = envpath.GetAppDir()+"/json/YoC.json"

func checkJsonDir()error{
	var dir,_ = envpath.GetParentDir(jsonPath)
	return envpath.CheckMakeDir(dir)
}

func JsonRead(device *map[string]*deviceStat) {
	path, err := jsonPath, checkJsonDir()
	if err != nil {
		Log.Fatal(err)
	}
	file, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
	if err != nil {
		Log.Println("warning:failed to load database file at: " + path)
		return
	}
	var scanner= bufio.NewReader(file)
	bytes, err := scanner.ReadBytes('\n')
	for err != io.EOF {
		if err != nil {
			Log.Fatal(err)
		}
		var dc dataClass
		err = json.Unmarshal(bytes, &dc)
		if err != nil {
			Log.Fatal(err)
		}
		(*device)[dc.deviceID] = new(deviceStat)
		(*device)[dc.deviceID].Data = &dc
	}
	defer func() {
		var err= file.Close()
		defer Log.Println(err)
	}()
}

func JsonWrite(device *map[string]*deviceStat) {
	path, err := jsonPath, checkJsonDir()
	if err != nil {
		Log.Fatal(err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 777)
	if err != nil {
		Log.Fatal(err)
	}
	var writer= bufio.NewWriter(file)
	for _, ds := range *device {
		dc, err := json.Marshal(ds)
		if err != nil {
			Log.Fatal(err)
		}
		_, err = writer.Write(dc)
		if err != nil {
			Log.Fatal(err)
		}
		err = writer.WriteByte('\n')
		if err != nil {
			Log.Fatal(err)
		}
	}
}