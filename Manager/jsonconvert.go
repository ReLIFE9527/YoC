package Data

import (
	"../Common"
	"../Log"
	"bufio"
	"encoding/json"
	"io"
	"os"
)

func JsonRead(device *map[string]*deviceStat) {
	var path= envpath.GetDBPath("YoC")
	var file, err= os.OpenFile(path, os.O_RDONLY, 777)
	if err != nil {
		YoCLog.Log.Println("warning:failed to load database file at: " + path)
		return
	}
	var scanner = bufio.NewReader(file)
	bytes, err := scanner.ReadBytes('\n')
	for err != io.EOF {
		if err != nil {
			YoCLog.Log.Fatal(err)
		}
		var dc dataClass
		err = json.Unmarshal(bytes, &dc)
		if err != nil {
			YoCLog.Log.Fatal(err)
		}
		(*device)[dc.deviceID] = new(deviceStat)
		(*device)[dc.deviceID].Data = &dc
	}
	defer func() {
		var err = file.Close()
		defer YoCLog.Log.Println(err)
	}()
}

func JsonWrite(device *map[string]*deviceStat) {
	var path = envpath.GetDBPath("YoC")
	var file, err= os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 777)
	if err != nil {
		YoCLog.Log.Fatal(err)
	}
	var writer = bufio.NewWriter(file)
	for _, ds := range *device {
		dc, err := json.Marshal(ds)
		if err != nil {
			YoCLog.Log.Fatal(err)
		}
		_, err = writer.Write(dc)
		if err != nil {
			YoCLog.Log.Fatal(err)
		}
		err = writer.WriteByte('\n')
		if err != nil {
			YoCLog.Log.Fatal(err)
		}
	}
}