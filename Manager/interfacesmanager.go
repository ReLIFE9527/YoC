package Data

import (
	"errors"
	"reflect"
	"time"
)

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
	var field =reflect.ValueOf(&obj).FieldByName(key)
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
	var field =reflect.ValueOf(&obj).FieldByName(key)
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

var saveChan chan string

func IMInit() error {
	devicesMap = make(map[string]*deviceOp)
	err := InitDevicesData()
	return err
}

func IMStart(ch *chan error) {
	var err error
	//TODO
	deviceStatUpt := make(chan string, 1)
	deviceStatUpt <- ""
	for true {
		select {
		case <-saveChan:
			err = DeviceSaveData()
			if err != nil {
				return
			}
		case <-deviceStatUpt:
			go func(ch chan string) {
				upt := make(chan string, 1)
				go func(ch chan string) {
					IMDeviceStatUpt()
					upt <- ""
				}(upt)
				time.Sleep(time.Second * 5)
				<-upt
				deviceStatUpt <- ""
			}(deviceStatUpt)
		default:
		}
	}
	defer func(e error) {
		*ch <- e
	}(err)
}

func IMShutDown() error{
	return nil
}

func IMDeviceLogin(device string) {
	devicesMap[device] = new(deviceOp{})
	DevicesOnline(device)
}

func IMDeviceLogout(device string) {
	DevicesOffline(device)
	delete(devicesMap, device)
}

func IMDeviceStatUpt() {
	for device,op:= range devicesMap {
		if op != nil {
			DeviceUpdate(device)
		}
	}
}

func IMDeviceRegister(device string,op string,fun interface{})error {
	err := devicesMap[device].Register(op, fun)
	return &IMError{device, "device register", err}
}

func IMClientLogin(client string) {
	clientsMap[client] = new(deviceOp{})
}

func IMClientLogout(client string) {
	delete(clientsMap, client)
}

func IMClientRegister(client string,op string,fun interface{})error {
	err := clientsMap[client].Register(op, fun)
	return &IMError{client, "client register", err}
}