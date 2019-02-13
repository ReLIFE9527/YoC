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

var collectorMap, gainerMap map[string]map[string]interface{}

func StorageInit() error {
	collectorMap = make(map[string]map[string]interface{})
	gainerMap = make(map[string]map[string]interface{})
	chanInit()
	err := initPassage()
	return err
}

func StorageStart(ch chan error) {
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
			err = passageSave()
			if err != nil {
				return
			}
		case <-chanMap["deviceUpt"]:
			go func(ch chan string) {
				upt := make(chan string, 1)
				go func(ch chan string) {
					statUpt()
					ch <- ""
				}(upt)
				time.Sleep(time.Second * 5)
				<-upt
				ch <- ""
			}(chanMap["deviceUpt"])
		case <-chanMap["remove"]:
			var remove = make(chan string, 1)
			go func(ch chan string) {
				removeCheck()
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

func StorageShutDown() error {
	err := passageSave()
	return err
}

const statUptTime = time.Second * 5
const saveUpt = int64(time.Hour / statUptTime)
const removeUpt = int64(time.Hour * 24 / statUptTime)

var statUptCount int64

func statUpt() {
	for device, op := range collectorMap {
		if op != nil {
			update(device)
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

func removeCheck() {
	removeOutDate()
}

func CollectorLogin(id, key string) {
	collectorMap[id] = make(map[string]interface{})
	online(id, key)
	chanMap["save"] <- ""
}

func CollectorLogout(device string) {
	offline(device)
	delete(collectorMap, device)
	chanMap["save"] <- ""
}

func CollectorRegister(addr string, operation string, function interface{}) {
	collectorMap[addr][operation] = function
}

func GainerLogin(client string) {
	gainerMap[client] = make(map[string]interface{})
}

func GainerLogout(client string) {
	delete(gainerMap, client)
}

func GainerRegister(addr string, operation string, function interface{}) {
	gainerMap[addr][operation] = function
}

func chanInit() {
	chanMap = make(map[string]chan string)
	for name, buffer := range selectChanList {
		chanMap[name] = make(chan string, buffer)
	}
}

func GetKey(id string) string {
	return key(id)
}

func GetOnlineList() (dst *map[string]bool) {
	onlineList(dst)
	return dst
}
