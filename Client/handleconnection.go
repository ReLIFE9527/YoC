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
	scanner     *bufio.Reader
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
	_ = conn.SetReadDeadline(time.Time{})
	var stream string
	//TODO
	for !cn.heartBreak {
		_ = conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
		stream, err = cn.scanner.ReadString(Pack.TailByte)
		for err == io.EOF && stream[len(stream)-1] != Pack.TailByte {
			var str string
			str, err = cn.scanner.ReadString(Pack.TailByte)
			stream += str
		}
		if len(stream) > 0 {
			cn.actionFresh <- ""
			stream, err = Pack.DePackString(stream)
			if Pack.IsStreamValid([]string{"operation"}, stream) {
				cn.dispatch(stream)
			}
		}
	}
	return err
}

func (cn *connection) clientAccessCheck(conn net.Conn) (err error) {
	const loginPassword, loginAccess, loginFail = "{\"login\":\"password\"}", "{\"login\":\"access\"}", "{\"login\":\"failed\"}"
	err = writeRepeat(conn, time.Second*2, []byte(Pack.PackString(loginPassword)))
	if err != nil {
		return err
	}
	var access, newAccess = make(chan string, 1), false
	go cn.clientVerify(conn, access, &newAccess)
	select {
	case cn.addr = <-access:
		if cn.addr != "" {
			if newAccess {
				key, _ := json.Marshal(map[string]string{"key": cn.addr})
				err = writeRepeat(conn, time.Second*2, []byte(Pack.PackString(string(key))))
				if err != nil {
					return io.EOF
				}
			}
			err = writeRepeat(conn, time.Second*2, []byte(Pack.PackString(loginAccess)))
			return err
		} else {
			err = writeRepeat(conn, time.Second*2, []byte(Pack.PackString(loginFail)))
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

func (cn *connection) clientVerify(conn net.Conn, ch chan string, re *bool) {
	_ = conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
	cn.scanner = bufio.NewReader(conn)
	bytes, _ := cn.scanner.ReadString(Pack.TailByte)
	for len(ch) == 0 {
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
								key := sha1.Sum([]byte(time.Now().String()))
								*re = true
								ch <- fmt.Sprintf("%x", key)
							}
						} else {
							ch <- ""
						}
					}
				}
			}
		}
		_ = conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
		bytes, _ = cn.scanner.ReadString(Pack.TailByte)
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

func (cn *connection) clientLogout() { Data.IMClientLogout(cn.addr) }

func (cn connection) dispatch(operation string) {
	fmt.Println(operation)
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
