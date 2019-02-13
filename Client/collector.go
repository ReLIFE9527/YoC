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
	*Connector
	id        string
	key       string
	operation string
	working   chan string
}

func (collector *Collector) loop() {

}

func (collector *Collector) extraInit() {
	collector.working = make(chan string, 1)
}

func (collector *Collector) preAction() {

}

func (collector *Collector) postAction() {

}

func (collector *Collector) checkAccess() error {
	const loginFailed, loginDone Pack.Stream = "{\"login\":\"failed\"}", "{\"login\":\"done\"}"
	var access = make(chan string, 1)
	go collector.verify(access)
	select {
	case key := <-access:
		if key != "" {
			stream := Pack.Convert2Stream(&map[string]string{"key": key})
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
		_ = collector.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
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
						collector.key = Data.GetKey(id)
						if key, ok := (*table)["key"]; ok && collector.key != "" {
							ch <- key
						} else {
							key = fmt.Sprintf("%x", sha1.Sum([]byte(time.Now().String())))
							ch <- key
							//override the key
							//if collector.key=="" {
							collector.key = key
							//}
						}
					} else {
						ch <- ""
					}
				}
			}
		}
	}
	_ = collector.conn.SetReadDeadline(time.Time{})
}
