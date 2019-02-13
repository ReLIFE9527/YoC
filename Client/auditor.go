package Client

import (
	. "../Log"
	"net"
)

type Auditor struct {
	listener         net.Listener
	network, address string
}

func (auditor *Auditor) Init() error {
	auditor.network = "udp"
	auditor.address = "localhost:12345"
	err := auditor.subInit()
	auditor.listener, err = net.Listen(auditor.network, auditor.address)
	if err == nil {
		Log.Println(auditor.network, auditor.address, "Waiting for connection...")
	}
	return err
}

func (auditor *Auditor) subInit() error {
	return nil
}

func (auditor *Auditor) Listen(errCh chan error) {
	for len(errCh) == 0 {
		conn, err := auditor.listener.Accept()
		if err != nil {
			errCh <- err
		} else {
			go auditor.handle(conn)
		}
	}
	defer func() {
		err := auditor.listener.Close()
		Log.Println(err)
		Log.Println(auditor.network, auditor.address, "listener closed.")
	}()
}

func (auditor *Auditor) handle(conn net.Conn) {
	return
}
