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
	"sync"
	"time"
)

type connection struct {
	actionRefresh  chan string
	addr           string
	beatBreak      bool
	conn           net.Conn
	scanner        *bufio.Reader
	working        sync.Mutex
	writeOperation string
}

var writeOp = make(map[string][]string)
var writeMutex = make(map[string]*sync.Mutex)

func (cn *connection) handleConnection(conn net.Conn) (err error) {
	cn.scanner, cn.conn = bufio.NewReader(conn), conn
	addr, err := loginProgress(conn)
	if err != nil {
		Log.Println(conn.RemoteAddr().String(), " : ", err)
	} else {
		_ = conn.Close()
		return nil
	}
	login(addr)
	go connectionHeartBeats(&cn.beatBreak, cn.actionRefresh)
	//TODO
	for {
		if len(writeOp[addr]) > 0 {
			writeMutex[addr].Lock()
			op := writeOp[addr][0]
			writeOp[addr] = writeOp[addr][1:]
			writeMutex[addr].Unlock()
			err = dispatchOp(op, conn)
			cn.actionRefresh <- ""
		} else {
			var buffer = make([]byte, 128)
			err = conn.SetReadDeadline(time.Now().Add(time.Microsecond * 50))
			if err != nil {
				Log.Println(err)
			}
			n, err := conn.Read(buffer)
			if err == nil && n > 0 {
				data := string(buffer[:n])
				err = dispatchRead(data, conn)
				cn.actionRefresh <- ""
			} else {
				if n == -1 {
					//break
					fmt.Println(err)
				}
				switch err {
				case io.EOF:
					cn.beatBreak = true
				}
			}
		}
		if cn.beatBreak {
			break
		}
	}
	defer func() {
		logout(addr)
		err = conn.Close()
		if err != nil {
			Log.Println(err)
		} else {
			Log.Println(addr, " : connection closed")
		}
	}()
	return err
}

func (cn *connection) deviceAccessCheck() (err error) {
	const loginStart, loginFailed, loginDone = "{\"login\":\"start\"}", "{\"login\":\"failed\"}", "{\"login\":\"done\"}"
	return err
}

func (cn *connection) deviceLogin() (err error) {
	err = cn.deviceAccessCheck()
	if err != nil {
		return err
	}
	Data.IMDeviceLogin(cn.addr)
	Log.Println(cn.addr, " : device connected")
	return err
}

func dispatchOp(str string, conn net.Conn) (err error) {
	return err
}

func dispatchRead(str string, conn net.Conn) (err error) {
	fmt.Println(str)
	return err
}

func login(device string) {
	writeOp[device] = make([]string, 0)
	writeMutex[device] = new(sync.Mutex)
	IMDeviceLogin(device)
	basicOperationRegister(device)
	Log.Println(device, " : connected")
}

func logout(device string) {
	IMDeviceLogout(device)
	writeMutex[device] = nil
	writeOp[device] = nil
}

func basicOperationRegister(device string) {
	err := IMDeviceRegister(device, "OpList", writeOp[device])
	if err != nil {
		Log.Println(err)
	}
	err = IMDeviceRegister(device, "WriteMutex", writeMutex[device])
	if err != nil {
		Log.Println(err)
	}
}

func loginProgress(conn net.Conn) (device string, err error) {
	var loginDone, loginStart, loginFail = "{\"login\":\"done\"}", "{\"login\":\"start\"}", "{\"login\":\"failed\"}"
	err = writeRepeat(conn, time.Second*2, []byte(Pack.PackString(loginStart)))
	if err != nil {
		return "", err
	}
	var loginCh, ID = make(chan string, 1), ""
	go func() {
		ID = loginVerify(loginCh, conn)
	}()
	select {
	case device = <-loginCh:
		err = nil
		if device == "" {
			key := sha1.Sum([]byte(ID))
			device = fmt.Sprintf("%x", key)
			var ret, _ = json.Marshal(map[string]string{"key": device})
			ret = []byte(Pack.PackString(string(ret)))
			_ = conn.SetWriteDeadline(time.Now().Add(+time.Millisecond * 10))
			_, err = conn.Write(ret)
		} else {
			key := sha1.Sum([]byte(ID))
			if device == fmt.Sprintf("%x", key) {
			} else {
				err = writeRepeat(conn, time.Second*2, []byte(Pack.PackString(loginFail)))
				if err != nil {
					return "", err
				}
				device = fmt.Sprintf("%x", key)
				var ret, _ = json.Marshal(map[string]string{"key": device})
				err = writeRepeat(conn, time.Second*2, []byte(Pack.PackString(string(ret))))
				if err != nil {
					return "", err
				}
			}
		}
	case <-time.After(time.Second * 10):
		loginCh <- ""
		device = ""
		err = nil
	}
	err = writeRepeat(conn, time.Second*2, []byte(Pack.PackString(loginDone)))
	if err != nil {
		return "", io.EOF
	}
	return device, err
}

func loginVerify(ch chan string, conn net.Conn) (id string) {
	_ = conn.SetReadDeadline(time.Now().Add(time.Millisecond * 10))
	scanner := bufio.NewReader(conn)
	bytes, err := scanner.ReadString(Pack.TailByte)
	for err != nil && err != io.EOF && len(ch) == 0 {
		if len(bytes) > 0 {
			str, err := Pack.DePackString(bytes)
			if err != nil {
				Log.Println(err)
			} else {
				if Pack.IsStreamValid([]string{"id"}, str) {
					dataMap := Pack.Convert2Map(str)
					id = (*dataMap)["id"]
					if key, ok := (*dataMap)["key"]; ok {
						ch <- key
					} else {
						ch <- ""
					}
				}
			}
		}
	}
	return id
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
