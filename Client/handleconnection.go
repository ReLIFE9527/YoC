package Client

import (
	. "../Log"
	"../Manager"
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
	addr        string
	actionFresh chan string
	heartBreak  bool
}

func (cn *connection) handleConnection(conn net.Conn) (err error) {
	defer func() {
		_ = conn.Close()
	}()
	err = cn.clientLogin(conn)
	if err != nil {
		return err
	}
	defer cn.clientLogout()
	cn.actionFresh = make(chan string, 1)
	cn.heartBreak = false
	go connectionHeartBeats(&cn.heartBreak, cn.actionFresh)
	//TODO
	for {
		if cn.heartBreak {
			break
		}
		scanner := bufio.NewReader(conn)
		str, err := scanner.ReadString(Pack.PackTailByte)
		if err != nil && err != io.EOF {
			return err
		}
		if len(str) > 0 {
			str, err = Pack.DePackString(str)
			if Pack.IsStreamValid([]string{"operation"}, str) {
				cn.dispatch(str)
			}
		}
	}
	return err
}

func (cn *connection) clientAccessCheck(conn net.Conn) (err error) {
	const loginPassword, loginAccess, loginFail = "{\"login\":\"password\"}", "{\"login\":\"access\"}", "{\"login\":\"failed\"}"
	err = writeRepeat(conn, time.Second*2, []byte(loginPassword))
	if err != nil {
		return err
	}
	var access = make(chan string, 1)
	go cn.clientVerify(conn, access)
	select {
	case cn.addr = <-access:
		if cn.addr != "" {
			err = writeRepeat(conn, time.Second*2, []byte(loginAccess))
			return err
		} else {
			err = writeRepeat(conn, time.Second*2, []byte(loginFail))
			return io.EOF
		}
	case <-time.After(time.Second * 10):
		access <- ""
		go func() {
			<-time.After(time.Second)
			<-access
		}()
		return io.EOF
	}
}

func (cn *connection) clientVerify(conn net.Conn, ch chan string) {
	_ = conn.SetReadDeadline(time.Now().Add(time.Millisecond * 10))
	scanner := bufio.NewReader(conn)
	bytes, err := scanner.ReadString(Pack.PackTailByte)
	for err != nil && err != io.EOF && len(ch) == 0 {
		if len(bytes) > 0 {
			str, err := Pack.DePackString(bytes)
			if err != nil {
				Log.Println(err)
			} else {
				if Pack.IsStreamValid([]string{"password"}, str) {
					var dataMap = make(map[string]string)
					err = json.Unmarshal([]byte(str), &dataMap)
					if err != nil {
						Log.Println(err)
					} else {
						if dataMap["password"] == clientPassword {
							if key, ok := dataMap["key"]; ok && key != "" {
								ch <- key
							} else {
								tempt := sha1.Sum([]byte(time.Now().String()))
								ch <- string(tempt[:])
							}
						} else {
							ch <- ""
						}
					}
				}
			}
		}
		bytes, err = scanner.ReadString(Pack.PackTailByte)
	}
}

func (cn *connection) clientLogin(conn net.Conn) (err error) {
	err = cn.clientAccessCheck(conn)
	if err != nil {
		return err
	}
	Data.IMClientLogin(cn.addr)
	Log.Println(cn.addr, " : client connected")
	return err
}

func (cn *connection) clientLogout() { Data.IMDeviceLogout(cn.addr) }

func (cn connection) dispatch(operation string) {
	fmt.Println(operation)
}

func writeRepeat(conn net.Conn, t time.Duration, data []byte) (err error) {
	var ch = make(chan string)
	go func() {
		_ = conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 10))
		_, err = conn.Write(data)

		for err != nil {
			_ = conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 10))
			_, err = conn.Write(data)
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
