package Data

import (
	. "../Log"
	"time"
)

type stat struct {
	isOnline bool
	Data     *repository
}

var coStats map[string]*stat

func initPassage() error {
	coStats = make(map[string]*stat)
	err := JsonRead(&coStats)
	if err != nil && !IsJsonEmpty(err) {
		return err
	}
	return nil
}

func passageSave() error {
	err := JsonWrite(&coStats)
	return err
}

func online(id, key string) {
	if coStats[id] == nil {
		coStats[id] = new(stat)
		coStats[id].Data = new(repository)
	}
	coStats[id].isOnline = true
	coStats[id].Data.ID = id
	coStats[id].Data.Key = key
	coStats[id].Data.LastLogin = time.Now()
}

func offline(device string) {
	coStats[device].isOnline = false
	coStats[device].Data.LastLogin = time.Now()
}

func update(id string) {
	if coStats[id] != nil {
		coStats[id].Data.LastLogin = time.Now()
	} else {
		Log.Println("can not find target id : ", id)
	}
}

func removeOutDate() {
	for i, device := range coStats {
		if !device.isOnline {
			var t1, t2 = time.Now(), device.Data.LastLogin.AddDate(0, 0, 15)
			if (t1.YearDay() > t2.YearDay() && t1.Year() == t2.YearDay()) || t1.Year() > t2.Year() {
				coStats[i] = nil
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
	for device, stat := range coStats {
		if stat.isOnline {
			(*can)[device] = true
		}
	}
}

func key(id string) string {
	if value, ok := coStats[id]; ok {
		return value.Data.Key
	} else {
		return ""
	}
}
