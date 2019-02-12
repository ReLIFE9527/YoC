package Data

import (
	"time"
)

var selectChanList = map[string]int64{ //List of [function name]channel buffer
	"deviceUpt": 1,
	"remove":    1,
	"save":      10,
}
var chanMap map[string]chan string

type IMError struct {
	Obj string
	Op  string
	Err error
}

func (e *IMError) Error() string {
	return e.Obj + " {" + e.Op + ": " + e.Err.Error() + "}"
}

var devicesMap, clientsMap map[string]map[string]interface{}

func IMInit() error {
	devicesMap = make(map[string]map[string]interface{})
	clientsMap = make(map[string]map[string]interface{})
	imChanInit()
	err := initDevicesData()
	return err
}

func IMStart(ch chan error) {
	var err error
	//TODO
	chanMap["deviceUpt"] <- ""
	chanMap["remove"] <- ""
	for true {
		select {
		case <-chanMap["save"]:
			//log.Println(len(saveChan))
			for len(chanMap["save"]) > 0 {
				<-chanMap["save"]
			}
			err = deviceSaveData()
			if err != nil {
				return
			}
		case <-chanMap["deviceUpt"]:
			go func(ch chan string) {
				upt := make(chan string, 1)
				go func(ch chan string) {
					imDeviceStatUpt()
					ch <- ""
				}(upt)
				time.Sleep(time.Second * 5)
				<-upt
				ch <- ""
			}(chanMap["deviceUpt"])
		case <-chanMap["remove"]:
			var remove = make(chan string, 1)
			go func(ch chan string) {
				imDeviceRemoveCheck()
				ch <- ""
			}(remove)
			<-remove
		default:
			time.Sleep(time.Microsecond)
		}
	}
	defer func(e error) {
		ch <- e
	}(err)
}

func IMShutDown() error {
	return nil
}

const statUptTime = time.Second * 5
const saveUpt = int64(time.Hour / statUptTime)
const removeUpt = int64(time.Hour * 24 / statUptTime)

var statUptCount int64

func imDeviceStatUpt() {
	for device, op := range devicesMap {
		if op != nil {
			deviceUpdate(device)
		}
	}
	statUptCount++
	if statUptCount%saveUpt == 0 {
		chanMap["save"] <- ""
	}
	if statUptCount > removeUpt {
		chanMap["remove"] <- ""
		statUptCount = 0
	}

}

func imDeviceRemoveCheck() {
	deviceRemoveOutDate()
}

func IMDeviceLogin(device string) {
	devicesMap[device] = make(map[string]interface{})
	devicesOnline(device)
	chanMap["save"] <- ""
}

func IMDeviceLogout(device string) {
	devicesOffline(device)
	delete(devicesMap, device)
	chanMap["save"] <- ""
}

func IMDeviceRegister(addr string, operation string, function interface{}) {
	devicesMap[addr][operation] = function
}

func IMClientLogin(client string) {
	clientsMap[client] = make(map[string]interface{})
}

func IMClientLogout(client string) {
	delete(clientsMap, client)
}

func IMClientRegister(addr string, operation string, function interface{}) {
	clientsMap[addr][operation] = function
}

func imChanInit() {
	chanMap = make(map[string]chan string)
	for name, buffer := range selectChanList {
		chanMap[name] = make(chan string, buffer)
	}
}
