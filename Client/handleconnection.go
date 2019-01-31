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
	actionRefresh chan string
	addr          string
	conn          net.Conn
	heartBreak    bool
	scanner       *bufio.Reader
}

func (cn *connection) handleConnection(conn net.Conn) (err error) {
	defer func() {
		_ = conn.Close()
	}()
	cn.scanner, cn.conn = bufio.NewReader(conn), conn
	err = cn.clientLogin()
	if err != nil {
		return err
	}
	defer cn.clientLogout()
	cn.actionRefresh = make(chan string, 1)
	cn.heartBreak = false
	go connectionHeartBeats(&cn.heartBreak, cn.actionRefresh)
	var stream string
	_ = cn.conn.SetReadDeadline(time.Time{})
	//TODO
	for !cn.heartBreak {
		stream, err = cn.scanner.ReadString(Pack.TailByte)
		for err == io.EOF && stream[len(stream)-1] != Pack.TailByte {
			var str string
			str, err = cn.scanner.ReadString(Pack.TailByte)
			stream += str
		}
		if len(stream) > 0 {
			cn.actionRefresh <- ""
			stream, err = Pack.DePackString(stream)
			if Pack.IsStreamValid([]string{"operation"}, stream) {
				cn.dispatch(stream)
			}
		}
	}
	return err
}

func (cn *connection) clientAccessCheck() (err error) {
	const loginPassword, loginAccess, loginFail = "{\"login\":\"password\"}", "{\"login\":\"access\"}", "{\"login\":\"failed\"}"
	err = writeRepeat(cn.conn, time.Second*2, []byte(Pack.PackString(loginPassword)))
	if err != nil {
		return err
	}
	var access = make(chan string, 1)
	go cn.clientVerify(access)
	defer func() {
		access <- ""
		time.Sleep(time.Second)
		<-access
	}()
	select {
	case cn.addr = <-access:
		if cn.addr != "" {
			err = writeRepeat(cn.conn, time.Second*2, []byte(Pack.PackString(loginAccess)))
			return err
		} else {
			err = writeRepeat(cn.conn, time.Second*2, []byte(Pack.PackString(loginFail)))
			return io.EOF
		}
	case <-time.After(time.Second * 10):
		return io.EOF
	}
}

func (cn *connection) clientVerify(ch chan string) {
	_ = cn.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
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
							key := sha1.Sum([]byte(time.Now().String()))
							ch <- fmt.Sprintf("%x", key)
						} else {
							ch <- ""
						}
					}
				}
			}
		}
		_ = cn.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
		bytes, _ = cn.scanner.ReadString(Pack.TailByte)
	}
}

func (cn *connection) clientLogin() (err error) {
	err = cn.clientAccessCheck()
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

func connectionHeartBeats(flag *bool, actionRefresh chan string) {
	for {
		select {
		case <-actionRefresh:
			if *flag {
				return
			}
		case <-time.After(time.Minute):
			*flag = true
			break
		}
	}
}
