package Device

import (
	. "../Log"
	"../Manager"
	. "../Manager"
	"../Pack"
	"bufio"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"
)

type connection struct {
	actionRefresh chan string
	addr          string
	beatBreak     bool
	conn          net.Conn
	operation     string
	scanner       *bufio.Reader
	working       chan string
}

func (cn *connection) handleConnection(conn net.Conn) (err error) {
	defer func() {
		_ = conn.Close()
	}()
	cn.scanner, cn.conn, cn.working = bufio.NewReader(conn), conn, make(chan string, 1)
	err = cn.deviceLogin()
	if err != nil {
		return err
	}
	defer cn.deviceLogout()
	cn.actionRefresh = make(chan string)
	cn.beatBreak = false
	go connectionHeartBeats(&cn.beatBreak, cn.actionRefresh)
	var stream string
	//TODO
	for !cn.beatBreak {
		if cn.operation != "" {
			cn.operationDispatch()
		} else {
			_ = cn.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
			stream, err = cn.scanner.ReadString(Pack.TailByte)
			if len(stream) > 0 {
				cn.actionRefresh <- ""
				stream, err = Pack.DePackString(stream)
				/*if Pack.IsStreamValid([]string{"operation"}, stream) {
					cn.streamDispatch(stream)
				}*/
				cn.streamDispatch(stream)
			}
		}
	}
	return err
}

func (cn *connection) deviceVerify(ch chan string, returnKey *bool) {
	_ = cn.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
	bytes, _ := cn.scanner.ReadString(Pack.TailByte)
	for len(ch) == 0 {
		if len(bytes) > 0 {
			str, err := Pack.DePackString(bytes)
			if err != nil {
				Log.Println(err)
			} else {
				if Pack.IsStreamValid([]string{"id"}, str) {
					var dataMap = make(map[string]string)
					err = json.Unmarshal([]byte(str), &dataMap)
					if err != nil {
						Log.Println(err)
					} else {
						sum := fmt.Sprintf("%x", sha1.Sum([]byte(time.Now().String())))
						if key, ok := dataMap["key"]; key == sum && ok {
						} else {
							*returnKey = true
						}
						ch <- sum
					}
				}
			}
		}
		_ = cn.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
		bytes, _ = cn.scanner.ReadString(Pack.TailByte)
	}
}

func (cn *connection) deviceAccessCheck() (err error) {
	const loginFailed, loginDone = "{\"login\":\"failed\"}", "{\"login\":\"done\"}"
	var access, returnKey = make(chan string, 1), false
	go cn.deviceVerify(access, &returnKey)
	defer func() {
		access <- ""
		time.Sleep(time.Second)
		<-access
	}()
	select {
	case cn.addr = <-access:
		if cn.addr != "" {
			if returnKey {
				ret, err := json.Marshal(map[string]string{"key": cn.addr})
				if err != nil {
					return err
				}
				err = writeRepeat(cn.conn, time.Second*2, []byte(Pack.PackString(string(ret[:]))))
				if err != nil {
					return io.EOF
				}
			}
			err = writeRepeat(cn.conn, time.Second*2, []byte(Pack.PackString(loginDone)))
			return err
		} else {
			err = writeRepeat(cn.conn, time.Second*2, []byte(Pack.PackString(loginFailed)))
			return io.EOF
		}
	case <-time.After(time.Second * 10):
		return io.EOF
	}
}

func (cn *connection) deviceLogin() (err error) {
	err = cn.deviceAccessCheck()
	if err != nil {
		return err
	}
	Data.IMDeviceLogin(cn.addr)
	Data.IMDeviceRegister(cn.addr, "add", cn.AddOperation)
	Log.Println(cn.addr, " : device connected")
	return err
}

func (cn *connection) deviceLogout() {
	IMDeviceLogout(cn.addr)
}

func (cn *connection) operationDispatch() {
	fmt.Println(cn.operation)
	cn.operation = ""
	<-cn.working
}

func (cn *connection) AddOperation(operation string) error {
	if cn.operation == "" && len(cn.working) == 0 {
		cn.working <- ""
		cn.operation = operation
		return nil
	} else {
		return io.EOF
	}
}

func (cn *connection) streamDispatch(stream string) {
	fmt.Println(stream)
}

func writeRepeat(conn net.Conn, t time.Duration, data []byte) (err error) {
	var ch = make(chan string, 1)
	go func() {
		_ = conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 100))
		_, err = conn.Write(data)
		var count int
		for err != nil && count < 2 {
			_ = conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 100))
			_, err = conn.Write(data)
			count++
		}
		ch <- ""
	}()
	select {
	case <-ch:
		return nil
	case <-time.After(t):
		return io.EOF
	}
}

func connectionHeartBeats(flag *bool, actionFresh chan string) {
	for {
		select {
		case <-actionFresh:
			if *flag {
				return
			}
		case <-time.After(time.Minute):
			*flag = true
			break
		}
	}
}
