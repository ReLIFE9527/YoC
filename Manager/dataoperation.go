package Data

import (
	. "../Log"
	"time"
)

type deviceStat struct {
	isOnline bool
	Data     *dataClass
}

func (obj deviceStat) Online(){
	obj.isOnline = true
}

func (obj deviceStat) Offline(){
	obj.isOnline=false
}

func (obj deviceStat) Stat() bool {
	return obj.isOnline
}

func (obj deviceStat) SetStat(stat bool) {
	if stat {
		obj.Online()
	} else {
		obj.Offline()
	}
}

var devices map[string]*deviceStat

func initDevicesData() error {
	devices = make(map[string]*deviceStat)
	err := JsonRead(&devices)
	if err != nil && !IsJsonEmpty(err) {
		return err
	}
	return nil
}

func deviceSaveData()error {
	err :=JsonWrite(&devices)
	return err
}

func devicesOnline(device string) {
	if devices[device] == nil {
		devices[device] = new(deviceStat)
		devices[device].Data = new(dataClass)
	}
	devices[device].Online()
	devices[device].Data.SetDeviceID(device)
	err := devices[device].Data.Set("LastLogin", time.Now().String())
	Log.Println(err)
}

func devicesOffline(device string) {
	devices[device].Offline()
	err := devices[device].Data.Set("LastLogin", time.Now().String())
	Log.Println(err)
}

func deviceUpdate(device string){
	err := devices[device].Data.Set("LastLogin", time.Now().String())
	Log.Println(err)
}

func deviceRemoveOutDate(){

}