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
	"strings"
	"time"
)

type connection struct {
	actionRefresh chan string
	addr          string
	conn          net.Conn
	heartBreak    bool
	readWriter    *bufio.ReadWriter
}

func (cn *connection) handleConnection(conn net.Conn) (err error) {
	defer func() {
		_ = conn.Close()
	}()
	cn.readWriter, cn.conn = bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), conn
	err = cn.clientLogin()
	if err != nil {
		return err
	}
	defer cn.clientLogout()
	cn.actionRefresh = make(chan string, 1)
	cn.heartBreak = false
	go connectionHeartBeats(&cn.heartBreak, cn.actionRefresh)
	var packet Pack.Packet
	var stream Pack.Stream
	var str string
	_ = cn.conn.SetReadDeadline(time.Time{})
	//TODO
	for !cn.heartBreak {
		str, err = cn.readWriter.ReadString(Pack.TailByte)
		packet = Pack.Packet(str)
		if len(packet) > 0 {
			cn.actionRefresh <- ""
			stream, err = Pack.DePackString(packet)
			if Pack.IsStreamValid([]string{"operation"}, stream) {
				n := cn.readWriter.Reader.Buffered()
				_, _ = cn.readWriter.Discard(n)
				cn.dispatch(stream)
			}
		}
	}
	return err
}

func (cn *connection) clientAccessCheck() (err error) {
	const loginPassword, loginAccess, loginFail = "{\"login\":\"password\"}", "{\"login\":\"access\"}", "{\"login\":\"failed\"}"
	err = cn.writeRepeat(Pack.StreamPack(loginPassword), time.Second*2)
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
			err = cn.writeRepeat(Pack.StreamPack(loginAccess), time.Second*2)
			cn.addr = cn.conn.RemoteAddr().String()
			return err
		} else {
			err = cn.writeRepeat(Pack.StreamPack(loginFail), time.Second*2)
			return io.EOF
		}
	case <-time.After(time.Second * 10):
		return io.EOF
	}
}

func (cn *connection) clientVerify(ch chan string) {
	_ = cn.conn.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
	bytes, _ := cn.readWriter.ReadString(Pack.TailByte)
	packet := Pack.Packet(bytes)
	for len(ch) == 0 {
		if len(packet) > 0 {
			str, err := Pack.DePackString(packet)
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
		bytes, _ = cn.readWriter.ReadString(Pack.TailByte)
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

func (cn connection) dispatch(operation Pack.Stream) {
	fmt.Println(operation)
	var data = Pack.Convert2Map(operation)
	if op, ok := (*data)["operation"]; ok {
		switch op {
		case "getOnline":
			cn.getOnline()
		default:
			Log.Println("operation invalid : ", op)
			fmt.Println("operation invalid : ", op)
		}
	}
}

func (cn connection) writeRepeat(packet Pack.Packet, t time.Duration) (err error) {
	var ch = make(chan string, 1)
	go func() {
		_, _ = cn.readWriter.WriteString(string(packet))
		err = cn.readWriter.Flush()
		if err != nil {
			Log.Println(err)
		}
		var count int
	repeat:
		_ = cn.conn.SetReadDeadline(time.Now().Add(t))
		str, err := cn.readWriter.ReadString(Pack.TailByte)
		stream, _ := Pack.DePackString(packet)
		if strings.Contains(str, "done") && Pack.IsStreamValid([]string{"read"}, stream) {
			count += 2
		}
		count++
		if err != nil && err != io.EOF && count < 3 && len(ch) < 1 {
			goto repeat
		}
		ch <- ""
	}()
	select {
	case <-ch:
		_ = cn.conn.SetReadDeadline(time.Time{})
		return nil
	case <-time.After(t):
		ch <- ""
		_ = cn.conn.SetReadDeadline(time.Time{})
		return io.EOF
	}
}

func (cn *connection) getOnline() {
	var online = Data.GetOnlineList()
	var stream Pack.Stream = "{"
	for device := range *online {
		stream += Pack.BuildBlock("device", device)
	}
	stream += "}"
	if !Pack.IsStreamValid([]string{"device"}, stream) {
		fmt.Println("stream format error : ", stream)
	}
	packet := Pack.StreamPack(stream)
	err := cn.writeRepeat(packet, time.Second)
	if err != nil {
		Log.Println(err)
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
