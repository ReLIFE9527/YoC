package Client

import (
	"../Data"
	. "../Debug"
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
				DebugLogger.Println(packet, err)
			} else {
				if Pack.IsStreamValid(stream, []string{"operation"}) {
					//TODO
				}
				if Pack.IsStreamValid(stream, []string{"test"}) {
					collector.testReceiver(stream)
				}
				if Pack.IsStreamValid(stream, []string{"stat"}) {
					collector.refreshLink(stream)
				}
			}
		}
	}
}

func (collector *Collector) extraInit() {
	collector.working = make(chan string, 1)
}

func (collector *Collector) preAction() {
	Data.CollectorLogin(collector.id, collector.key)
	DebugLogger.Println(collector.addr, " : collector connected")
	DebugLogger.Println("id : ", collector.id)
}

func (collector *Collector) postAction() {
	Data.CollectorLogout(collector.id)
	DebugLogger.Println(collector.addr, " : collector disconnected")
}

func (collector *Collector) checkAccess() error {
	const loginFailed, loginDone, loginFailed2 Pack.Stream = "{\"login\":\"failed\"}", "{\"login\":\"done\"}", "{\"login\":\"ID is used by other devices\"}"
	var access = make(chan string, 1)
	go collector.verify(access)
	select {
	case key := <-access:
		switch key {
		case "nil":
			stream := Pack.Convert2Stream(&map[string]string{"key": collector.key})
			packet := Pack.StreamPack(stream)
			err := collector.writeRepeat(packet, time.Second)
			if err != nil {
				return io.EOF
			}
			err = collector.writeRepeat(Pack.StreamPack(loginDone), time.Second)
			return err
		case "no key":
			_ = collector.writeRepeat(Pack.StreamPack(loginFailed2), time.Second)
			return io.EOF
		case collector.key:
			err := collector.writeRepeat(Pack.StreamPack(loginDone), time.Second)
			return err
		default:
			_ = collector.writeRepeat(Pack.StreamPack(loginFailed), time.Second)
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

func (collector *Collector) verify(ch chan string) {
	var bytes string
	for len(ch) == 0 {
		bytes, _ = collector.readWriter.ReadString(Pack.TailByte)
		packet := Pack.Packet(bytes)
		if len(packet) > 0 {
			stream, err := Pack.DePack(packet)
			if err != nil {
				DebugLogger.Println(err)
			} else {
				if Pack.IsStreamValid(stream, []string{"id"}) {
					var table = Pack.Convert2Map(stream)
					if id, ok := (*table)["id"]; ok {
						collector.id = id
						collector.key = Data.GetKey(id)
						if key, ok := (*table)["key"]; ok && collector.key != "" {
							ch <- key
							return
						} else {
							key = fmt.Sprintf("%x", sha1.Sum([]byte(time.Now().String())))
							if collector.key == "" {
								//override the key
								collector.key = key
								ch <- "nil"
							} else {
								ch <- "no key"
							}
							return
						}
					}
				}
			}
		}
	}
}
