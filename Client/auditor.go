package Client

import (
	. "../Log"
	"net"
)

type Functions interface {
	subInit() error
	handle(net.Conn)
	open()
	listen(chan error)
}

type Auditor struct {
	Functions
}

type auditor struct {
	Functions
	listener         net.Listener
	address, network string
	conn             Connector
}

func (auditor *Auditor) Init(f Functions) error {
	auditor.Functions = f
	err := auditor.subInit()
	if err == nil {
		auditor.open()
	}
	return err
}

func (auditor *Auditor) Listen(errCh chan error) {
	auditor.listen(errCh)
}

func (auditor *auditor) subInit() error {
	auditor.network = "udp"
	auditor.address = "localhost:12345"
	return nil
}

func (auditor *auditor) handle(conn net.Conn) {
	return
}

func (auditor *auditor) open() {
	var err error
	auditor.listener, err = net.Listen(auditor.network, auditor.address)
	if err == nil {
		Log.Println(auditor.network, auditor.address, "Waiting for connection...")
	}
}

func (auditor *auditor) listen(errCh chan error) {
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

type Auditor32375 struct {
	auditor
}

func (auditor *Auditor32375) subInit() error {
	auditor.network = "tcp"
	auditor.address = "localhost:32375"
	return nil
}

func (auditor *Auditor32375) handle(conn net.Conn) {
	auditor.conn = new(Collector)
	err := auditor.conn.Handle(conn)
	if err != nil {
		Log.Println(err)
	}
}

type Auditor32376 struct {
	auditor
	Password string
}

func (auditor *Auditor32376) subInit() error {
	auditor.network = "tcp"
	auditor.address = "localhost:32376"
	return nil
}

func (auditor *Auditor32376) handle(conn net.Conn) {
	auditor.conn = &Gainer{password: auditor.Password}
	err := auditor.conn.Handle(conn)
	if err != nil {
		Log.Println(err)
	}
}
