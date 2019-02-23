package Client

import (
	. "../Debug"
	"net"
)

type auditorFunc interface {
	subInit() error
	handle(net.Conn)
	open()
	listen(chan error, chan net.Conn)
}

type Auditor struct {
	auditorFunc
}

type auditor struct {
	auditorFunc
	listener         net.Listener
	address, network string
	conn             Connector
}

func (auditor *Auditor) Init(f auditorFunc) error {
	auditor.auditorFunc = f
	err := auditor.subInit()
	if err == nil {
		auditor.open()
	}
	return err
}

func (auditor *Auditor) Listen(errCh chan error) {
	var handle = make(chan net.Conn)
	go auditor.listen(errCh, handle)
	for len(errCh) == 0 {
		select {
		case conn := <-handle:
			auditor.handle(conn)
			//case <-time.After(time.Millisecond * 100):
		}
	}
}

func (auditor *auditor) subInit() error {
	auditor.network = "udp"
	auditor.address = "localhost:12345"
	return nil
}

func (auditor *auditor) handle(net.Conn) {
	return
}

func (auditor *auditor) open() {
	var err error
	auditor.listener, err = net.Listen(auditor.network, auditor.address)
	if err == nil {
		DebugLogger.Println(auditor.network, auditor.address, "Waiting for connection...")
	}
}

func (auditor *auditor) listen(errCh chan error, handle chan net.Conn) {
	for len(errCh) == 0 {
		conn, err := auditor.listener.Accept()
		if err != nil {
			errCh <- err
		} else {
			handle <- conn
		}
	}
	defer func() {
		err := auditor.listener.Close()
		DebugLogger.Println(err)
		DebugLogger.Println(auditor.network, auditor.address, "listener closed.")
	}()
}

type Auditor32375 struct {
	auditor
}

func (auditor *Auditor32375) subInit() error {
	auditor.network = "tcp"
	auditor.address = ":32375"
	return nil
}

func (auditor *Auditor32375) handle(conn net.Conn) {
	auditor.conn.Init(new(Collector))
	err := auditor.conn.Handle(conn)
	if err != nil {
		DebugLogger.Println(err)
	}
}

type Auditor32376 struct {
	auditor
	Password string
}

func (auditor *Auditor32376) subInit() error {
	auditor.network = "tcp"
	auditor.address = ":32376"
	return nil
}

func (auditor *Auditor32376) handle(conn net.Conn) {
	auditor.conn.Init(&Gainer{password: auditor.Password})
	err := auditor.conn.Handle(conn)
	if err != nil {
		DebugLogger.Println(err)
	}
}
