package Client

import (
	"../Data"
	. "../Log"
	"../Pack"
	"encoding/json"
	"io"
	"time"
)

type Gainer struct {
	connector
	password string
}

func (gainer *Gainer) switcher(stream Pack.Stream) {
	var table = Pack.Convert2Map(stream)
	if action, ok := (*table)["operation"]; ok {
		gainer.refresh <- ""
		switch action {
		case "getOnlineList":
			gainer.getOnline()
		default:
			Log.Println("unknown operation : ", action)
		}
	}
}

func (gainer *Gainer) loop() {
	var str, _ = gainer.readWriter.ReadString(Pack.TailByte)
	var packet = Pack.Packet(str)
	if len(packet) > 0 {
		gainer.refresh <- ""
		stream, err := Pack.DePack(packet)
		if err != nil {
			Log.Println(err)
		} else {
			if Pack.IsStreamValid(stream, []string{"operation"}) {
				//TODO
			}
			if Pack.IsStreamValid(stream, []string{"test"}) {
				gainer.testReceiver(stream)
			}
		}
	}
}

func (gainer *Gainer) checkAccess() error {
	const loginPassword, loginAccess, loginFail Pack.Stream = "{\"login\":\"password\"}", "{\"login\":\"access\"}", "{\"login\":\"failed\"}"
	err := gainer.writeRepeat(Pack.StreamPack(loginPassword), time.Second)
	if err != nil {
		return err
	}
	var access = make(chan string, 1)
	go gainer.verify(access)
	select {
	case stat := <-access:
		if stat == "success" {
			err = gainer.writeRepeat(Pack.StreamPack(loginAccess), time.Second)
			return err
		} else {
			_ = gainer.writeRepeat(Pack.StreamPack(loginFail), time.Second)
			return io.EOF
		}
	case <-time.After(time.Second * 10):
		go func() {
			access <- ""
			time.Sleep(time.Second)
			<-access
		}()
		return io.EOF
	}
}

func (gainer *Gainer) verify(ch chan string) {
	var bytes string
	for len(ch) == 0 {
		bytes, _ = gainer.readWriter.ReadString(Pack.TailByte)
		packet := Pack.Packet(bytes)
		if len(packet) > 0 {
			stream, err := Pack.DePack(packet)
			if err != nil {
				Log.Println(err)
			} else {
				if Pack.IsStreamValid(stream, []string{"password"}) {
					var dataMap = make(map[string]string)
					err = json.Unmarshal([]byte(stream), &dataMap)
					if err != nil {
						Log.Println(err)
					} else {
						if dataMap["password"] == gainer.password {
							ch <- "success"
						} else {
							ch <- "fail"
						}
					}
				}
			}
		}
	}
}

func (gainer *Gainer) preAction() {
	Data.GainerLogin(gainer.addr)
	Log.Println(gainer.addr, " : gainer connected")
}

func (gainer *Gainer) postAction() {
	Data.GainerLogout(gainer.addr)
	Log.Println(gainer.addr, " : gainer disconnected")
}

func (gainer *Gainer) getOnline() {
	var online = Data.GetOnlineList()
	var blocks string
	for device := range *online {
		blocks += Pack.BuildBlock("device", device)
	}
	var stream = Pack.Blocks2Stream(blocks)
	packet := Pack.StreamPack(stream)
	err := gainer.writeRepeat(packet, time.Second)
	if err != nil {
		Log.Println(err)
	}
}
