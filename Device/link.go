package Device

import (
	. "../Log"
	"net"
)

var listener net.Listener

func LinkInit() (err error) {
	listener, err = net.Listen("tcp", "localhost:32376")
	if err != nil {
		Log.Println(err)
	} else {
		Log.Println("Waiting for devices connection...")
	}
	return err
}

func LinkHandle(ch chan error) {
	for len(ch) < 1 {
		conn, err := listener.Accept()
		if err != nil {
			Log.Println(err)
			ch <- err
		} else {
			go func(conn net.Conn) {
				var cn connection
				var err = cn.handleConnection(conn)
				if err != nil {
					Log.Println(err)
				}
			}(conn)
		}
	}
	defer func() {
		err := listener.Close()
		if err != nil {
			Log.Println(err)
			ch <- err
		} else {
			Log.Println("devices listener closed.")
		}
	}()
}
