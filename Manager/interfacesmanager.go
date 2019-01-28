package Data

import (
	"errors"
	"reflect"
	"time"
)

var selectChanList  = map[string]int64{ //List of [function name]channel buffer
	"deviceUpt": 1,
	"remove":    1,
	"save":      10,
}

type IMError struct {
	Obj string
	Op string
	Err error
}
func (e *IMError) Error() string {
	return e.Obj + " " + e.Op + ": " + e.Err.Error()
}

type deviceOp struct {
}
func (obj *deviceOp)Register(key string,value interface{}) error {
	if key=="this" {
		return errors.New("dataClass write access error")
	}
	var field =reflect.ValueOf(obj).Elem().FieldByName(key)
	if !field.IsValid(){
		return errors.New("can't find target element")
	}
	if field.Type()==reflect.ValueOf(value).Type() {
		field.Set(reflect.ValueOf(value))
	}else{
		return errors.New("field value type error")
	}
	return nil
}
var devicesMap map[string]*deviceOp

type clientOp struct {
}
func (obj *clientOp)Register(key string,value interface{}) error {
	if key=="this" {
		return errors.New("dataClass write access error")
	}
	var field =reflect.ValueOf(obj).Elem().FieldByName(key)
	if !field.IsValid(){
		return errors.New("can't find target element")
	}
	if field.Type()==reflect.ValueOf(value).Type() {
		field.Set(reflect.ValueOf(value))
	}else{
		return errors.New("field value type error")
	}
	return nil
}
var clientsMap map[string]*clientOp

func IMInit() error {
	devicesMap = make(map[string]*deviceOp)
	imChanInit()
	err := initDevicesData()
	return err
}

func IMStart(ch chan error) {
	var err error
	//TODO
	chanMap["deviceUpt"] <- ""
	chanMap["remove"] <-""
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

func IMShutDown() error{
	return nil
}

const statUptTime = time.Second*5
const saveUpt = int64(time.Hour/statUptTime)
const removeUpt = int64(time.Hour*24/statUptTime)
var statUptCount int64

func imDeviceStatUpt() {
	for device, op := range devicesMap {
		if op != nil {
			deviceUpdate(device)
		}
	}
	statUptCount ++
	if statUptCount % saveUpt ==0 {
		chanMap["save"] <- ""
	}
	if statUptCount>removeUpt {
		chanMap["remove"] <- ""
		statUptCount = 0
	}

}

func imDeviceRemoveCheck() {
	deviceRemoveOutDate()
}

func IMDeviceLogin(device string) {
	devicesMap[device] = new(deviceOp)
	devicesOnline(device)
	chanMap["save"] <- ""
}

func IMDeviceLogout(device string) {
	devicesOffline(device)
	delete(devicesMap, device)
	chanMap["save"] <- ""
}

func IMDeviceRegister(device string,op string,fun interface{})error {
	err := devicesMap[device].Register(op, fun)
	return &IMError{device, "device register", err}
}

func IMClientLogin(client string) {
	clientsMap[client] = new(clientOp)
}

func IMClientLogout(client string) {
	delete(clientsMap, client)
}

func IMClientRegister(client string,op string,fun interface{})error {
	err := clientsMap[client].Register(op, fun)
	return &IMError{client, "client register", err}
}

var chanMap map[string]chan string
func imChanInit() {
	chanMap = make(map[string]chan string)
	for name,buffer := range selectChanList {
		chanMap[name] =  make(chan string,buffer)
	}
}
