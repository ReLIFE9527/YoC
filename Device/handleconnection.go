package Device

import (
	. "../Log"
	. "../Manager"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var writeOp = make(map[string][]string)
var writeMutex = make(map[string]*sync.Mutex)

func handleConnection(conn net.Conn) (err error) {
	var addr, beatBreak, beatFresh = conn.RemoteAddr().String(), false, make(chan string)
	login(addr)
	go connectionHeartBeats(&beatBreak, beatFresh)
	//TODO
	for {
		if len(writeOp[addr]) > 0 {
			writeMutex[addr].Lock()
			op := writeOp[addr][0]
			writeOp[addr] = writeOp[addr][1:]
			writeMutex[addr].Unlock()
			err = dispatchOp(op, conn)
			beatFresh <- ""
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
				beatFresh <- ""
			} else {
				if n == -1 {
					//break
					fmt.Println(err)
				}
				switch err {
				case io.EOF:
					beatBreak = true
				}
			}
		}
		if beatBreak {
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

func connectionHeartBeats(flag *bool, ch chan string) {
	for {
		select {
		case <-ch:
			if *flag {
				return
			}
		case <-time.After(time.Minute):
			*flag = true
			break
		}
	}
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
	baseOperationRegister(device)
	Log.Println(device, " : connected")
}

func logout(device string) {
	IMDeviceLogout(device)
	writeMutex[device] = nil
	writeOp[device] = nil
}

func baseOperationRegister(device string) {
	err := IMDeviceRegister(device, "OpList", writeOp[device])
	if err != nil {
		Log.Println(err)
	}
	err = IMDeviceRegister(device, "WriteMutex", writeMutex[device])
	if err != nil {
		Log.Println(err)
	}
}
