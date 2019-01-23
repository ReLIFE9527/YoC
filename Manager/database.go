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
	if (stat) {
		obj.Online()
	} else {
		obj.Offline()
	}
}

var Devices *map[string]*deviceStat

func InitDataBase() {
	Devices = new(map[string]*deviceStat)
	JsonRead(Devices)
}

