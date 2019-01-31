package Client

import . "../Log"
import "net"

var listener net.Listener
var clientPassword string

func LinkInit(password string) (err error) {
	clientPassword = password
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
				var err = handleConnection(conn)
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
