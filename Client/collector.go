package Client

import (
	"../Data"
	. "../Log"
	"../Pack"
	"crypto/sha1"
	"fmt"
	"io"
	"time"
)

type Collector struct {
	connector
	id        string
	key       string
	operation string
	working   chan string
}

func (collector *Collector) loop() {
	if collector.operation != "" {
		//TODO
	} else {
		str, _ := collector.readWriter.ReadString(Pack.TailByte)
		packet := Pack.Packet(str)
		if len(packet) > 0 {
			stream, err := Pack.DePack(packet)
			if err != nil {
				return
			}
			fmt.Println(stream)
			if Pack.IsStreamValid(stream, []string{"operation"}) {
				//TODO
			}
			if Pack.IsStreamValid(stream, []string{"test"}) {
				collector.testReceiver(stream)
			}
		}
	}
}

func (collector *Collector) extraInit() {
	collector.working = make(chan string, 1)
}

func (collector *Collector) preAction() {
	Data.CollectorLogin(collector.id, collector.key)
	Log.Println(collector.addr, " : collector connected")
	Log.Println("id : ", collector.id)
}

func (collector *Collector) postAction() {
	Data.CollectorLogout(collector.id)
	Log.Println(collector.addr, " : collector disconnected")
}

func (collector *Collector) checkAccess() error {
	const loginFailed, loginDone Pack.Stream = "{\"login\":\"failed\"}", "{\"login\":\"done\"}"
	var access = make(chan string, 1)
	go collector.verify(access)
	select {
	case key := <-access:
		if key == "nil" {
			stream := Pack.Convert2Stream(&map[string]string{"key": collector.key})
			packet := Pack.StreamPack(stream)
			err := collector.writeRepeat(packet, time.Second)
			if err != nil {
				return io.EOF
			}
			err = collector.writeRepeat(Pack.StreamPack(loginDone), time.Second)
			return err
		} else {
			if key == collector.key {
				err := collector.writeRepeat(Pack.StreamPack(loginDone), time.Second)
				return err
			} else {
				_ = collector.writeRepeat(Pack.StreamPack(loginFailed), time.Second)
				return io.EOF
			}
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

func (collector *Collector) verify(ch chan string) {
	var bytes string
	for len(ch) == 0 {
		bytes, _ = collector.readWriter.ReadString(Pack.TailByte)
		packet := Pack.Packet(bytes)
		if len(packet) > 0 {
			stream, err := Pack.DePack(packet)
			if err != nil {
				Log.Println(err)
			} else {
				if Pack.IsStreamValid(stream, []string{"id"}) {
					var table = Pack.Convert2Map(stream)
					if id, ok := (*table)["id"]; ok {
						collector.id = id
						collector.key = Data.GetKey(id)
						if key, ok := (*table)["key"]; ok && collector.key != "" {
							ch <- key
						} else {
							key = fmt.Sprintf("%x", sha1.Sum([]byte(time.Now().String())))
							ch <- "nil"
							//override the key
							//if collector.key=="" {
							collector.key = key
							//}
						}
					}
				}
			}
		}
	}
}
