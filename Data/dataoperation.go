package Data

import (
	. "../Log"
	"time"
)

type deviceStat struct {
	isOnline bool
	Data     *repository
}

func (obj deviceStat) Online() {
	obj.isOnline = true
}

func (obj deviceStat) Offline() {
	obj.isOnline = false
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

func deviceSaveData() error {
	err := JsonWrite(&devices)
	return err
}

func devicesOnline(device string) {
	if devices[device] == nil {
		devices[device] = new(deviceStat)
		devices[device].Data = new(repository)
	}
	devices[device].Online()
	devices[device].Data.SetDeviceID(device)
	err := devices[device].Data.Set("LastLogin", time.Now())
	if err != nil {
		Log.Println(err)
	}
}

func devicesOffline(device string) {
	devices[device].Offline()
	err := devices[device].Data.Set("LastLogin", time.Now())
	if err != nil {
		Log.Println(err)
	}
}

func deviceUpdate(device string) {
	if devices[device] != nil {

		err := devices[device].Data.Set("LastLogin", time.Now())
		if err != nil {
			Log.Println(err)
		}
	} else {
		devicesOnline(device)
	}
}

func deviceRemoveOutDate() {
	for i, device := range devices {
		if !device.isOnline {
			var t1, t2 = time.Now(), device.Data.LastLogin.AddDate(0, 0, 15)
			//var s1, s2= t1.String(), t2.String()
			//Log.Println(s1 + "\n" + s2)
			if (t1.YearDay() > t2.YearDay() && t1.Year() == t2.YearDay()) || t1.Year() > t2.Year() {
				devices[i] = nil
			}
		}
	}
	err := deviceSaveData()
	if err != nil {
		Log.Println(err)
	}
}

func GetOnlineList() *map[string]bool {
	online := new(map[string]bool)
	for device, stat := range devices {
		if stat.isOnline {
			(*online)[device] = true
		}
	}
	return online
}
