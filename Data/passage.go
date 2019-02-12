package Data

import (
	. "../Log"
	"time"
)

type stat struct {
	isOnline bool
	Data     *repository
}

func (passage stat) Online() {
	passage.isOnline = true
}

func (passage stat) Offline() {
	passage.isOnline = false
}

func (passage stat) Stat() bool {
	return passage.isOnline
}

func (passage stat) SetStat(stat bool) {
	if stat {
		passage.Online()
	} else {
		passage.Offline()
	}
}

var devices map[string]*stat

func initPassage() error {
	devices = make(map[string]*stat)
	err := JsonRead(&devices)
	if err != nil && !IsJsonEmpty(err) {
		return err
	}
	return nil
}

func passageSave() error {
	err := JsonWrite(&devices)
	return err
}

func online(id, key string) {
	if devices[id] == nil {
		devices[id] = new(stat)
		devices[id].Data = new(repository)
	}
	devices[id].Online()
	devices[id].Data.ID = id
	devices[id].Data.Key = key
	devices[id].Data.LastLogin = time.Now()
}

func offline(device string) {
	devices[device].Offline()
	devices[device].Data.LastLogin = time.Now()
}

func update(id string) {
	if devices[id] != nil {
		devices[id].Data.LastLogin = time.Now()
	} else {
	}
}

func removeOutDate() {
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
	err := passageSave()
	if err != nil {
		Log.Println(err)
	}
}

func onlineList(can *map[string]bool) {
	can = new(map[string]bool)
	for device, stat := range devices {
		if stat.isOnline {
			(*can)[device] = true
		}
	}
}

func key(id string) string {
	return devices[id].Data.Key
}
