package Device

import (
	. "../Log"
	. "../Manager"
	"fmt"
	"net"
)

func handleConnection(conn net.Conn) (err error) {
	var addr = conn.RemoteAddr().String()
	IMDeviceLogin(addr)
	Log.Println(addr," : connected")
	//TODO
	for {
		var buffer= make([]byte, 128)
		n, err := conn.Read(buffer)
		if err == nil && n > 0 {
			data := string(buffer[:n])
			fmt.Println(data)
		}
		if false {
			break
		}
	}
	defer func() {
		IMDeviceLogout(addr)
		err = conn.Close()
		if err != nil {
			Log.Println(err)
		} else {
			Log.Println("connection closed :", addr)
		}
	}()
	return err
}