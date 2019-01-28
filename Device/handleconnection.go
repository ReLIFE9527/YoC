package Device

import (
	. "../Log"
	. "../Manager"
	"net"
)

func handleConnection(conn net.Conn) (err error) {
	var addr= conn.RemoteAddr().String()
	IMDeviceLogin(addr)
	//TODO
	for {
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