package Data

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

var devices *map[string]*deviceStat

func InitDevicesData() error {
	devices = new(map[string]*deviceStat)
	err := JsonRead(devices)
	if err != nil && !IsJsonEmpty(err) {
		return err
	}
	return nil
}