package Client

import (
	. "../Log"
	"../Manager"
	"../Pack"
	"bufio"
	"crypto/sha1"
	"encoding/json"
	"io"
	"net"
	"time"
)

func handleConnection(conn net.Conn) (err error) {
	var addr string
	defer func() {
		_ = conn.Close()
	}()
	err = clientLogin(conn, &addr)
	if err != nil {
		return err
	}
	//TODO

	return err
}

func clientAccessCheck(conn net.Conn) (addr string, err error) {
	const loginPassword, loginAccess, loginFail = "{\"login\":\"password\"}", "{\"login\":\"access\"}", "{\"login\":\"failed\"}"
	err = writeRepeat(conn, time.Second*2, []byte(loginPassword))
	if err != nil {
		return "", err
	}
	var access = make(chan string, 1)
	go clientVerify(conn, access)
	select {
	case addr = <-access:
		if addr != "" {
			err = writeRepeat(conn, time.Second*2, []byte(loginAccess))
			return addr, err
		} else {
			err = writeRepeat(conn, time.Second*2, []byte(loginFail))
			return "", io.EOF
		}
	case <-time.After(time.Second * 10):
		access <- ""
		go func() {
			<-time.After(time.Second)
			<-access
		}()
		return "", io.EOF
	}
}

func clientVerify(conn net.Conn, ch chan string) {
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
								ch <- string(sha1.Sum([]byte(time.Now().String()))[:])
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

func clientLogin(conn net.Conn, addr *string) (err error) {
	*addr, err = clientAccessCheck(conn)
	if err != nil {
		return err
	}
	Data.IMClientLogin(*addr)
	Log.Println(*addr, " : client connected")
	return err
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
